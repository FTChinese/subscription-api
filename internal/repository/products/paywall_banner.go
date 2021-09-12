package products

import "github.com/FTChinese/subscription-api/pkg/pw"

// retrieveBanner retrieves a banner and the optional promo attached to it.
// The banner id is fixed to 1.
func (env Env) retrieveBanner() (pw.BannerSchema, error) {
	var schema pw.BannerSchema

	err := env.dbs.Read.Get(&schema, pw.StmtBanner)
	if err != nil {
		return pw.BannerSchema{}, err
	}

	return schema, nil
}

type bannerResult struct {
	value pw.BannerSchema
	error error
}

// asyncRetrieveBanner retrieves banner in a goroutine.
func (env Env) asyncRetrieveBanner() <-chan bannerResult {
	c := make(chan bannerResult)

	go func() {
		defer close(c)

		banner, err := env.retrieveBanner()

		c <- bannerResult{
			value: banner,
			error: err,
		}
	}()

	return c
}
