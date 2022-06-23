package blockexplorer

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

var (
	//go:embed postgres/schema/00-cosmos.sql
	schema00Cosmos string

	//go:embed postgres/schema/01-auth.sql
	schema01Auth string

	//go:embed postgres/schema/02-bank.sql
	schema02Bank string

	//go:embed postgres/schema/03-staking.sql
	schema03Staking string

	//go:embed postgres/schema/04-consensus.sql
	schema04Consensus string

	//go:embed postgres/schema/05-mint.sql
	schema05Mint string

	//go:embed postgres/schema/06-distribution.sql
	schema06Distribution string

	//go:embed postgres/schema/07-pricefeed.sql
	schema07PriceFeed string

	//go:embed postgres/schema/08-gov.sql
	schema08Gov string

	//go:embed postgres/schema/09-modules.sql
	schema09Modules string

	//go:embed postgres/schema/10-slashing.sql
	schema10Slashing string

	//go:embed postgres/schema/11-feegrant.sql
	schema11FeeGrant string
)

// LoadPostgresSchema loads schema required by block explorer into postgres database
func LoadPostgresSchema(ctx context.Context, db *pgx.Conn) error {
	for _, cmds := range []string{
		schema00Cosmos,
		schema01Auth,
		schema02Bank,
		schema03Staking,
		schema04Consensus,
		schema05Mint,
		schema06Distribution,
		schema07PriceFeed,
		schema08Gov,
		schema09Modules,
		schema10Slashing,
		schema11FeeGrant,
	} {
		if _, err := db.Exec(ctx, cmds); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
