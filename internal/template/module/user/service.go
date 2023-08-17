package user

import (
	"context"
	"net/http"

	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/constant"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/data/sqlxpkg"
	"github.com/tanveerprottoy/stdlib-go-template/internal/template/module/user/dto"
	"github.com/tanveerprottoy/stdlib-go-template/internal/template/module/user/entity"
	"github.com/tanveerprottoy/stdlib-go-template/pkg/errorpkg"
	"github.com/tanveerprottoy/stdlib-go-template/pkg/timepkg"
)

type Service struct {
	repository sqlxpkg.Repository[entity.User]
}

func NewService(r sqlxpkg.Repository[entity.User]) *Service {
	s := new(Service)
	s.repository = r
	return s
}

func (s *Service) ReadOneInternal(id string) (entity.User, error) {
	return s.repository.ReadOne(id)
}

func (s *Service) Create(d *dto.CreateUpdateUserDTO, ctx context.Context) (entity.User, *errorpkg.HTTPError) {
	// convert dto to entity
	e := entity.User{}
	e.Name = d.Name
	// b.Role = d.Role
	n := timepkg.NowUnixMilli()
	e.CreatedAt = n
	e.UpdatedAt = n
	err := s.repository.Create(&e)
	if err != nil {
		return e, errorpkg.MakeDBError(err)
	}
	return e, nil
}

func (s *Service) ReadMany(limit, page int, ctx context.Context) (map[string]any, *errorpkg.HTTPError) {
	m := make(map[string]any)
	m["items"] = make([]entity.User, 0)
	m["limit"] = limit
	m["page"] = page
	offset := limit * (page - 1)
	d, err := s.repository.ReadMany(limit, offset)
	if err != nil {
		return m, errorpkg.MakeDBError(err)
	}
	m["items"] = d
	return m, nil
}

func (s *Service) ReadOne(id string, ctx context.Context) (entity.User, *errorpkg.HTTPError) {
	b, err := s.ReadOneInternal(id)
	if err != nil {
		return b, errorpkg.MakeDBError(err)
	}
	return b, nil
}

func (s *Service) Update(id string, d *dto.CreateUpdateUserDTO, ctx context.Context) (entity.User, *errorpkg.HTTPError) {
	b, err := s.ReadOneInternal(id)
	if err != nil {
		return b, errorpkg.MakeDBError(err)
	}
	b.Name = d.Name
	b.UpdatedAt = timepkg.NowUnixMilli()
	rows, err := s.repository.Update(id, &b)
	if err != nil {
		return b, errorpkg.MakeDBError(err)
	}
	if rows > 0 {
		return b, nil
	}
	return b, &errorpkg.HTTPError{Code: http.StatusBadRequest, Err: errorpkg.NewError(constant.OperationNotSuccess)}
}

func (s *Service) Delete(id string, ctx context.Context) (entity.User, *errorpkg.HTTPError) {
	b, err := s.ReadOneInternal(id)
	if err != nil {
		return b, errorpkg.MakeDBError(err)
	}
	rows, err := s.repository.Delete(id)
	if err != nil {
		return b, errorpkg.MakeDBError(err)
	}
	if rows > 0 {
		return b, nil
	}
	return b, &errorpkg.HTTPError{Code: http.StatusBadRequest, Err: errorpkg.NewError(constant.OperationNotSuccess)}
}
