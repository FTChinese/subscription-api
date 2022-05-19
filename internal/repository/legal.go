package repository

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/subscription-api/internal/pkg/legal"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

type LegalRepo struct {
	dbs    db.ReadWriteMyDBs
	logger *zap.Logger
}

func NewLegalRepo(dbs db.ReadWriteMyDBs, logger *zap.Logger) LegalRepo {
	return LegalRepo{
		dbs:    dbs,
		logger: logger,
	}
}

func (repo LegalRepo) Create(l legal.Legal) error {
	_, err := repo.dbs.Write.NamedExec(
		legal.StmtInsertLegal,
		l)

	return err
}

func (repo LegalRepo) UpdateContent(l legal.Legal) error {
	_, err := repo.dbs.Write.NamedExec(
		legal.StmtUpdateLegal,
		l)

	return err
}

func (repo LegalRepo) UpdateStatus(l legal.Legal) error {
	_, err := repo.dbs.Write.NamedExec(
		legal.StmtUpdateStatus,
		l)

	return err
}

func (repo LegalRepo) Retrieve(id string) (legal.Legal, error) {
	var l legal.Legal

	err := repo.dbs.Read.Get(
		&l,
		legal.StmtRetrieveLegal,
		id,
	)

	if err != nil {
		return legal.Legal{}, err
	}

	return l, nil
}

func (repo LegalRepo) countLegalRows() (int64, error) {
	var count int64
	err := repo.dbs.Read.Get(
		&count,
		legal.StmtCount,
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (repo LegalRepo) listLegal(p gorest.Pagination) ([]legal.Legal, error) {
	var list = make([]legal.Legal, 0)
	err := repo.dbs.Read.Select(
		&list,
		legal.StmtListLegal,
		p.Limit,
		p.Offset())

	if err != nil {
		return nil, err
	}

	return list, nil
}

func (repo LegalRepo) ListLegal(p gorest.Pagination) (legal.List, error) {
	defer repo.logger.Sync()
	sugar := repo.logger.Sugar()

	countCh := make(chan int64)
	listCh := make(chan legal.List)

	go func() {
		defer close(countCh)
		n, err := repo.countLegalRows()
		if err != nil {
			sugar.Error(err)
		}

		countCh <- n
	}()

	go func() {
		defer close(listCh)
		list, err := repo.listLegal(p)
		if err != nil {
			sugar.Error(err)
		}
		listCh <- legal.List{
			Total:      0,
			Pagination: gorest.Pagination{},
			Data:       list,
			Err:        err,
		}
	}()

	count, listResult := <-countCh, <-listCh

	if listResult.Err != nil {
		return legal.List{}, listResult.Err
	}

	listResult.Total = count
	listResult.Pagination = p

	return listResult, nil
}
