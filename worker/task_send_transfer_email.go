package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	log "github.com/rs/zerolog/log"
	"github.com/diatbbin/QuotaFlow/mail"
)

type PayloadSendTransferEmail struct {
	RecipientUsername string `json:"recipient_username"`
	SenderEmail       string `json:"sender_email"`
	Tokens            int64  `json:"tokens"`
}

const TaskSendTransferEmail = "task:send_transfer_email"

func (distributor *RedisTaskDistributor) DistributorTaskSendTransferEmail(
	ctx context.Context,
	payload *PayloadSendTransferEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}
	task := asynq.NewTask(TaskSendTransferEmail, jsonPayload, opts...)
	taskInfo, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().
		Str("type", taskInfo.Type).
		Bytes("payload", taskInfo.Payload).
		Int("max_retry", taskInfo.MaxRetry).
		Str("queue", taskInfo.Queue).
		Str("task_id", taskInfo.ID).
		Msg("enqueued transfer notification email task")

	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendTransferEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendTransferEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	recipient, err := processor.store.GetUser(ctx, payload.RecipientUsername)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("recipient user does not exist: %w", asynq.SkipRetry)
		}
		return fmt.Errorf("failed to get recipient user: %w", err)
	}

	subject := mail.TransferNotificationSubject()
	content := mail.TransferNotificationContent(payload.SenderEmail, payload.Tokens)

	err = processor.mailSender.SendEmail(
		subject,
		content,
		[]string{recipient.Email},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to send transfer email: %w", err)
	}

	log.Info().
		Str("type", task.Type()).
		Str("recipient_email", recipient.Email).
		Str("sender_email", payload.SenderEmail).
		Int64("tokens", payload.Tokens).
		Msg("sent transfer notification email")

	return nil
}
