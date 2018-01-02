package handler

import (
	"net/http"

	"openauth/api/exception"
	"openauth/api/server/http/context"
	"openauth/api/server/http/request"
	"openauth/api/server/http/response"
)

// TestDomainID Just for test
const TestDomainID = "fa735972-b059-44f3-b95f-78f0aaa1306e"

// CreateProject use to create an project
func CreateProject(w http.ResponseWriter, r *http.Request) {
	val, err := request.CheckObjectBody(r)
	if err != nil {
		response.Failed(w, err)
		return
	}

	// get did from token
	did := TestDomainID

	name := val.Get("name").ToString()
	desc := val.Get("description").ToString()

	if name == "" {
		response.Failed(w, err)
		return
	}

	// 交给业务控制层处理
	project, err := projectsrv.CreateProject(did, name, desc)
	if err != nil {
		response.Failed(w, err)
		return
	}

	response.Success(w, http.StatusCreated, project)
	return
}

// GetProject use to get one project
func GetProject(w http.ResponseWriter, r *http.Request) {
	ps := context.GetParamsFromContext(r)
	pid := ps.ByName("pid")

	// TODO: get token from context, and check permission
	proj, err := projectsrv.GetProject(pid)
	if err != nil {
		response.Failed(w, err)
		return
	}

	response.Success(w, http.StatusOK, proj)
	return
}

// ListProject use to list all project
func ListProject(w http.ResponseWriter, r *http.Request) {

	projects, err := projectsrv.ListProject(TestDomainID)
	if err != nil {
		response.Failed(w, err)
		return
	}

	response.Success(w, http.StatusOK, projects)
	return
}

// DeleteProject use to delete an project
func DeleteProject(w http.ResponseWriter, r *http.Request) {
	ps := context.GetParamsFromContext(r)
	pid := ps.ByName("pid")

	// TODO: get token from context, and check permission
	if err := projectsrv.DestroyProject(pid); err != nil {
		response.Failed(w, err)
		return
	}

	response.Success(w, http.StatusNoContent, "")
	return
}

// ListProjectUsers use to list
func ListProjectUsers(w http.ResponseWriter, r *http.Request) {
	ps := context.GetParamsFromContext(r)
	pid := ps.ByName("pid")

	users, err := projectsrv.ListProjectUsers(pid)
	if err != nil {
		response.Failed(w, err)
		return
	}

	response.Success(w, http.StatusOK, users)
	return
}

// AddUsersToProject add users
func AddUsersToProject(w http.ResponseWriter, r *http.Request) {
	ps := context.GetParamsFromContext(r)
	pid := ps.ByName("pid")

	iter, err := request.CheckArrayBody(r)
	if err != nil {
		response.Failed(w, err)
		return
	}

	// get did from token
	uids := []string{}
	for iter.ReadArray() {
		uids = append(uids, iter.ReadString())
	}
	if iter.Error != nil {
		response.Failed(w, exception.NewBadRequest("json format decode error, %s", iter.Error))
		return
	}

	if len(uids) == 0 {
		response.Failed(w, exception.NewBadRequest("not uid find"))
		return
	}

	// 业务层处理逻辑
	if err := projectsrv.AddUsersToProject(pid, uids...); err != nil {
		response.Failed(w, err)
		return
	}

	response.Success(w, http.StatusCreated, "")
	return

}

// RemoveUsersFromProject remove users
func RemoveUsersFromProject(w http.ResponseWriter, r *http.Request) {
	ps := context.GetParamsFromContext(r)
	pid := ps.ByName("pid")

	iter, err := request.CheckArrayBody(r)
	if err != nil {
		response.Failed(w, err)
		return
	}

	// get did from token
	uids := []string{}
	for iter.ReadArray() {
		uids = append(uids, iter.ReadString())
	}
	if iter.Error != nil {
		response.Failed(w, exception.NewBadRequest("json format decode error, %s", iter.Error))
		return
	}

	if len(uids) == 0 {
		response.Failed(w, exception.NewBadRequest("not uid find"))
		return
	}

	// 业务逻辑层
	if err := projectsrv.RemoveUsersFromProject(pid, uids...); err != nil {
		response.Failed(w, err)
		return
	}

	response.Success(w, http.StatusCreated, "")
	return
}