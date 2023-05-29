package repository

import (
	"log"

	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/price"
)

// countPrices gets the total entry of stripe prices
// in db.
func (repo StripeRepo) countPrices(live bool) (int64, error) {
	var count int64
	err := repo.dbs.Read.Get(
		&count,
		price.StmtCountStripePrice,
		live,
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (repo StripeRepo) listPrices(live bool, p gorest.Pagination) ([]price.StripePrice, error) {
	list := make([]price.StripePrice, 0)

	err := repo.dbs.Read.Select(
		&list,
		price.StmtStripePagedPrices,
		live,
		p.Limit,
		p.Offset(),
	)

	if err != nil {
		return nil, err
	}

	return list, nil
}

// ListPricesPages retrieves a list of stripe prices
// with pagination.
func (repo StripeRepo) ListPricesPaged(
	live bool,
	p gorest.Pagination,
) (pkg.PagedData[price.StripePrice], error) {

	countCh := make(chan int64)
	listCh := make(chan pkg.AsyncResult[[]price.StripePrice])

	go func() {
		defer close(countCh)
		n, err := repo.countPrices(live)
		if err != nil {
			log.Print(err)
		}

		countCh <- n
	}()

	go func() {
		defer close(listCh)
		list, err := repo.listPrices(live, p)
		listCh <- pkg.AsyncResult[[]price.StripePrice]{
			Err:   err,
			Value: list,
		}
	}()

	count, listResult := <-countCh, <-listCh

	if listResult.Err != nil {
		return pkg.PagedData[price.StripePrice]{}, listResult.Err
	}

	return pkg.PagedData[price.StripePrice]{
		Total:      count,
		Pagination: p,
		Data:       listResult.Value,
	}, nil
}

// Retrieveprice retrieves a single stripe price.
func (repo StripeRepo) RetrievePrice(id string, live bool) (price.StripePrice, error) {
	var p price.StripePrice

	err := repo.dbs.Read.Get(
		&p,
		price.StmtStripePrice,
		id,
		live)

	if err != nil {
		return price.StripePrice{}, err
	}

	return p, nil
}

func (repo StripeRepo) UpsertPrice(p price.StripePrice) error {

	_, err := repo.dbs.Write.NamedExec(
		price.StmtUpsertStripePrice,
		p)

	if err != nil {
		return err
	}

	return nil
}

// ActivatePrice adds an entry into product_active_price table.
func (repo StripeRepo) ActivatePrice(p price.StripePrice) error {
	tx, err := txrepo.BeginPriceTx(repo.dbs.Write)
	if err != nil {
		return err
	}

	err = tx.UpsertActivePrice(p.ActiveEntry())
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (repo StripeRepo) DeactivePrice(p price.StripePrice) error {
	tx, err := txrepo.BeginPriceTx(repo.dbs.Delete)
	if err != nil {
		return err
	}

	err = tx.RemoveActivePrice(p.ActiveEntry())
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil

}
