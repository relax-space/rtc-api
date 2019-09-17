package factory

import (
	"context"

	"github.com/go-xorm/xorm"
	"github.com/pangpanglabs/goutils/echomiddleware"
)

func DB(ctx context.Context) xorm.Interface {
	v := ctx.Value(echomiddleware.ContextDBName)
	if v == nil {
		panic("DB is not exist")
	}
	if db, ok := v.(xorm.Interface); ok {
		return db
	}
	if db, ok := v.(*xorm.Engine); ok {
		return db.NewSession()
	}
	panic("DB is not exist")
}

