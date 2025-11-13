//go:build wireinject

package main

import (
	"github.com/google/wire"

	pagerepo "github.com/naka-sei/tsudzuri/infrastructure/db/page"
	ipostgres "github.com/naka-sei/tsudzuri/infrastructure/db/postgres"
	userrepo "github.com/naka-sei/tsudzuri/infrastructure/db/user"
	presentationgrpc "github.com/naka-sei/tsudzuri/presentation/grpc"
	grpcpage "github.com/naka-sei/tsudzuri/presentation/grpc/page"
	grpcuser "github.com/naka-sei/tsudzuri/presentation/grpc/user"
	pageusecase "github.com/naka-sei/tsudzuri/usecase/page"
	useservice "github.com/naka-sei/tsudzuri/usecase/service"
	userusecase "github.com/naka-sei/tsudzuri/usecase/user"
)

func InitializePresentationServer(
	dbConn *ipostgres.Connection,
) (*presentationgrpc.Server, error) {
	wire.Build(
		presentationSet,
		usecaseSet,
		repoSet,
		serviceSet,
	)
	return nil, nil
}

func transactionServiceProvider(dbc *ipostgres.Connection) useservice.TransactionService {
	return dbc
}

var (
	presentationSet = wire.NewSet(
		grpcpage.NewCreateService,
		grpcpage.NewGetService,
		grpcpage.NewListService,
		grpcpage.NewEditService,
		grpcpage.NewDeleteService,
		grpcpage.NewLinkAddService,
		grpcpage.NewLinkRemoveService,
		grpcpage.NewJoinService,
		grpcuser.NewCreateService,
		grpcuser.NewLoginService,
		presentationgrpc.NewServer,
	)
	usecaseSet = wire.NewSet(
		pageusecase.NewCreateUsecase,
		pageusecase.NewGetUsecase,
		pageusecase.NewListUsecase,
		pageusecase.NewEditUsecase,
		pageusecase.NewDeleteUsecase,
		pageusecase.NewLinkAddUsecase,
		pageusecase.NewLinkRemoveUsecase,
		pageusecase.NewJoinUsecase,
		userusecase.NewCreateUsecase,
		userusecase.NewLoginUsecase,
	)
	repoSet = wire.NewSet(
		pagerepo.NewPageRepository,
		userrepo.NewUserRepository,
	)
	serviceSet = wire.NewSet(
		transactionServiceProvider,
	)
)
