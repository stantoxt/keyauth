package store

import (
	"github.com/defineiot/keyauth/dao/models"
	"github.com/defineiot/keyauth/internal/exception"
)

// CreateMemberUser use to create user
func (s *Store) CreateMemberUser(u *models.User) error {
	// 判断用户名是否存在
	if _, err := s.dao.User.GetUser(models.AccountIndex, u.Account); err != nil {
		if _, ok := err.(*exception.NotFound); !ok {
			return err
		}
	} else {
		return exception.NewBadRequest("account: %s is exist", u.Account)
	}

	// 判断用户的秘密是否符合复杂度要求

	// 如果用户未选择部门, 则使用默认部门
	if u.Department.ID == "" {
		defaultDep, err := s.dao.Department.GetDepartmentByName(u.Domain.ID, defaultDepartmentName)
		if err != nil {
			return err
		}

		u.Department.ID = defaultDep.ID
	}

	// 创建用户
	if err := s.dao.User.CreateUser(u); err != nil {
		return err
	}

	// 查询出域的具体详情
	dom, err := s.dao.Domain.GetDomainByID(u.Domain.ID)
	if err != nil {
		return err
	}

	// 查询用户部门的详情
	dep, err := s.dao.Department.GetDepartment(u.Department.ID)
	if err != nil {
		return err
	}

	// 部门有允许加入的项目则加入项目
	projects, err := s.dao.Project.ListDepartmentProjects(dep.ID)
	if err != nil {
		return err
	}
	if len(projects) > 0 {
		if err := s.dao.User.AddProjectsToUser(dom.ID, u.ID, projects); err != nil {
			return err
		}
	}

	// 部门有相关角色则赋予相关人员
	roles, err := s.dao.Role.ListDepartmentRoles(dep.ID)
	if err != nil {
		return err
	}
	for _, r := range roles {
		if err := s.dao.User.BindRole(dom.ID, u.ID, r.ID); err != nil {
			return err
		}
	}

	u.Domain = dom
	u.Department = dep
	u.Roles = roles

	return nil
}

// ListMemberUsers list all user
func (s *Store) ListMemberUsers(domainID string) ([]*models.User, error) {
	users, err := s.dao.User.ListDomainUsers(domainID)
	if err != nil {
		return nil, err
	}

	for i := range users {
		u := users[i]
		// 查询出域的具体详情
		dom, err := s.dao.Domain.GetDomainByID(u.Domain.ID)
		if err != nil {
			return nil, err
		}

		// 查询用户部门的详情
		dep, err := s.dao.Department.GetDepartment(u.Department.ID)
		if err != nil {
			return nil, err
		}

		roles, err := s.dao.Role.ListUserRole(u.Domain.ID, u.ID)
		if err != nil {
			return nil, err
		}

		u.Domain = dom
		u.Department = dep
		u.Roles = roles
	}

	return users, nil
}

// GetUser get an user
func (s *Store) GetUser(domainID, userID string) (*models.User, error) {
	var err error

	u := new(models.User)
	cacheKey := s.cachePrefix.user + userID

	if s.isCache {
		if s.cache.Get(cacheKey, u) {
			s.log.Debug("get project from cache key: %s", cacheKey)
			return u, nil
		}
		s.log.Debug("get project from cache failed, key: %s", cacheKey)
	}

	u, err = s.dao.User.GetUser(models.UserIDIndex, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, exception.NewBadRequest("user %s not found", userID)
	}

	// 查询出域的具体详情
	dom, err := s.dao.Domain.GetDomainByID(u.Domain.ID)
	if err != nil {
		return nil, err
	}

	// 查询用户部门的详情
	dep, err := s.dao.Department.GetDepartment(u.Department.ID)
	if err != nil {
		return nil, err
	}

	// 查询用户的角色
	roles, err := s.dao.Role.ListUserRole(u.Domain.ID, u.ID)
	if err != nil {
		return nil, err
	}

	// 查询用户的项目
	projects, err := s.dao.Project.ListUserProjects(u.Domain.ID, u.ID)
	if err != nil {
		return nil, err
	}

	u.Domain = dom
	u.Department = dep
	u.Roles = roles
	u.Projects = projects

	// 查询用户的默认项目
	if u.DefaultProject.ID != "" {
		pro, err := s.dao.Project.GetProjectByID(u.DefaultProject.ID)
		if err != nil {
			return nil, err
		}
		u.DefaultProject = pro
	}

	if s.isCache {
		if !s.cache.Set(cacheKey, u, s.ttl) {
			s.log.Debug("set user cache failed, key: %s", cacheKey)
		}
		s.log.Debug("set user cache ok, key: %s", cacheKey)
	}

	return u, nil
}

// DeleteUser delete an user by id
func (s *Store) DeleteUser(domainID, userID string) error {
	cacheKey := s.cachePrefix.user + userID

	if err := s.dao.User.DeleteUser(domainID, userID); err != nil {
		return err
	}

	if s.isCache {
		if !s.cache.Delete(cacheKey) {
			s.log.Debug("delete user from cache failed, key: %s", cacheKey)
		}
		s.log.Debug("delete user from cache success, key: %s", cacheKey)
	}

	return nil
}

// ListProjectUser list all user
func (s *Store) ListProjectUser(projectID string) ([]*models.User, error) {
	users, err := s.dao.User.ListProjectUsers(projectID)
	if err != nil {
		return nil, err
	}

	for i := range users {
		u := users[i]
		// 查询出域的具体详情
		dom, err := s.dao.Domain.GetDomainByID(u.Domain.ID)
		if err != nil {
			return nil, err
		}

		// 查询用户部门的详情
		dep, err := s.dao.Department.GetDepartment(u.Department.ID)
		if err != nil {
			return nil, err
		}

		roles, err := s.dao.Role.ListUserRole(u.Domain.ID, u.ID)
		if err != nil {
			return nil, err
		}

		u.Domain = dom
		u.Department = dep
		u.Roles = roles
	}

	return users, nil
}

// BindRole todo
func (s *Store) BindRole(domainID, userID, roleName string) error {
	ok, err := s.dao.Role.CheckRoleExist(roleName)
	if err != nil {
		return err
	}
	if !ok {
		return exception.NewBadRequest("role: %s not exist", roleName)
	}

	cacheKey := "user_" + userID
	if s.isCache {
		if !s.cache.Delete(cacheKey) {
			s.log.Debug("delete user from cache failed, key: %s", cacheKey)
		}
		s.log.Debug("delete user from cache success, key: %s", cacheKey)
	}

	return s.dao.User.BindRole(domainID, userID, roleName)
}

// UnBindRole todo
func (s *Store) UnBindRole(domainID, userID, roleName string) error {
	ok, err := s.dao.Role.CheckRoleExist(roleName)
	if err != nil {
		return err
	}
	if !ok {
		return exception.NewBadRequest("role: %s not exist", roleName)
	}

	cacheKey := "user_" + userID
	if s.isCache {
		if !s.cache.Delete(cacheKey) {
			s.log.Debug("delete user from cache failed, key: %s", cacheKey)
		}
		s.log.Debug("delete user from cache success, key: %s", cacheKey)
	}

	return s.dao.User.UnBindRole(domainID, userID, roleName)
}
