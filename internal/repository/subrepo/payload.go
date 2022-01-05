package subrepo

import (
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/wechat"
)

func (env Env) SaveAliWebhookPayload(p ali.WebhookPayload) error {
	_, err := env.dbs.Write.NamedExec(
		ali.StmtSavePayload,
		p)

	if err != nil {
		return err
	}

	return nil
}

func (env Env) SaveAliOrderQueryPayload(p ali.OrderQueryPayload) error {
	_, err := env.dbs.Write.NamedExec(
		ali.StmtSavePayload,
		p)

	if err != nil {
		return err
	}

	return nil
}

func (env Env) SaveWxPayload(schema wechat.PayloadSchema) error {
	_, err := env.dbs.Write.NamedExec(
		wechat.StmtSavePayload,
		schema)

	if err != nil {
		return err
	}

	return nil
}
