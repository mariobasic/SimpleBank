package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	db "github.com/mariobasic/simplebank/db/sqlc"
	"github.com/mariobasic/simplebank/util"
	"github.com/rs/zerolog/log"
)

const TaskSendVerifyEmail = "task:send_verify_email"

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

func (r *RedisTaskDistributor) DistributeTaskSendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to jsonPayload payload: %w", err)
	}
	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)
	info, err := r.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().
		Str("type", task.Type()).
		Bytes("payload", jsonPayload).
		Str("queue", info.Queue).
		Int("max_retry", info.MaxRetry).
		Msg("enqueued task")

	return nil
}

func (r *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("cannot unmarshal payload: %w", asynq.SkipRetry)
	}

	user, err := r.store.GetUser(ctx, payload.Username)
	if err != nil {
		// if the db traffic is heavy this could be invoked before the user tx is commited
		// than we will not find the user in time, and we want retry to pick it up
		// another way to do this is to add async processing delay
		//if errors.Is(err, pgx.ErrNoRows) {
		//	return fmt.Errorf("user %s not found: %w", payload.Username, asynq.SkipRetry)
		//}
		return fmt.Errorf("cannot get user %s: %w", payload.Username, err)
	}

	verifyEmail, err := r.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: util.RandomString(32),
	})
	if err != nil {
		return fmt.Errorf("cannot create verify email: %w", err)
	}
	err = r.sendEmail(verifyEmail)
	if err != nil {
		return fmt.Errorf("cannot send verify email: %w", err)
	}

	log.Info().
		Str("task", task.Type()).
		Bytes("payload", task.Payload()).
		Str("email", user.Email).
		Msg("processed task")

	return nil
}

func (r *RedisTaskProcessor) sendEmail(verifyEmail db.VerifyEmail) error {
	subject := "Welcome to Simple Bank"
	verifyUrl := fmt.Sprintf("http://localhost:8080/v1/verify_email?email_id=%d&secret_code=%s", verifyEmail.ID, verifyEmail.SecretCode)
	content := fmt.Sprintf(`Hello %s,<br/>
	Thank you for registering with us!<br/>
	Please <a href="%s">click here</a> to verify your email address.<br/>`, verifyEmail.Username, verifyUrl)
	to := []string{verifyEmail.Email}

	return r.mailer.Send(subject, content, to, nil, nil, nil)
}
