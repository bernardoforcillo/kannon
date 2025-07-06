package adminapi

import (
	"context"

	sqlc "github.com/kannon-email/kannon/internal/db"
	"github.com/kannon-email/kannon/internal/domains"
	"github.com/kannon-email/kannon/internal/templates"

	pb "github.com/kannon-email/kannon/proto/kannon/admin/apiv1"
)

type adminAPIService struct {
	dm domains.DomainManager
	tm templates.Manager
}

func (s *adminAPIService) GetDomains(ctx context.Context, in *pb.GetDomainsReq) (*pb.GetDomainsResponse, error) {
	domains, err := s.dm.GetAllDomains(ctx)
	if err != nil {
		return nil, err
	}

	res := pb.GetDomainsResponse{}
	for _, domain := range domains {
		res.Domains = append(res.Domains, dbDomainToProtoDomain(domain))
	}
	return &res, nil
}

func (s *adminAPIService) GetDomain(ctx context.Context, in *pb.GetDomainReq) (*pb.GetDomainRes, error) {
	domain, err := s.dm.FindDomain(ctx, in.Domain)
	if err != nil {
		return nil, err
	}

	return &pb.GetDomainRes{
		Domain: dbDomainToProtoDomain(domain),
	}, nil
}

func (s *adminAPIService) CreateDomain(ctx context.Context, in *pb.CreateDomainRequest) (*pb.Domain, error) {
	domain, err := s.dm.CreateDomain(ctx, in.Domain)
	if err != nil {
		return nil, err
	}

	return dbDomainToProtoDomain(domain), nil
}

func (s *adminAPIService) RegenerateDomainKey(ctx context.Context, in *pb.RegenerateDomainKeyRequest) (*pb.Domain, error) {
	domain, err := s.dm.RegenerateDomainKey(ctx, in.Domain)
	if err != nil {
		return nil, err
	}

	return dbDomainToProtoDomain(domain), nil
}

func dbDomainToProtoDomain(in sqlc.Domain) *pb.Domain {
	return &pb.Domain{
		Domain:     in.Domain,
		Key:        in.Key,
		DkimPubKey: in.DkimPublicKey,
	}
}
