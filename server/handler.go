package main

import (
	"net/http"

	"github.com/knanao/goauth/server/model"
	"github.com/labstack/echo"
)

func setRoute(e *echo.Echo) {
	e.GET("/", handleIndexGet)
	e.GET("/login", handleLoginGet)
	e.POST("/login", handleLoginPost)
	e.POST("/logout", handleLogoutPost)
	e.GET("/users/:user_id", handleUsers)
	e.POST("/users/:user_id", handleUsers)

	admin := e.Group("/admin", MiddlewareAuthAdmin)
	admin.GET("", handleAdmin)
	admin.POST("", handleAdmin)
	admin.GET("/users", handleAdminUsersGet)
}

func handleIndexGet(c echo.Context) error {
	return c.Render(http.StatusOK, "index", "longin checker")
}

func handleUsers(c echo.Context) error {
	userID := c.Param("user_id")
	err := CheckUserID(c, userID)
	if err != nil {
		c.Echo().Logger.Debugf("User Page[%s] Role Error. [%s]", userID, err)
		msg := "You have not logged in."
		return c.Render(http.StatusOK, "error", msg)
	}
	users, err := userDA.FindByUserID(c.Param("user_id"), model.FindFirst)
	if err != nil {
		return c.Render(http.StatusOK, "error", err)
	}
	user := users[0]
	return c.Render(http.StatusOK, "user", user)
}

func handleAdmin(c echo.Context) error {
	return c.Render(http.StatusOK, "admin", nil)
}

func handleAdminUsersGet(c echo.Context) error {
	users, err := userDA.FindAll()
	if err != nil {
		return c.Render(http.StatusOK, "error", err)
	}
	return c.Render(http.StatusOK, "admin_users", users)
}

func handleLoginGet(c echo.Context) error {
	return c.Render(http.StatusOK, "login", nil)
}

func handleLoginPost(c echo.Context) error {
	userID := c.FormValue("userid")
	password := c.FormValue("password")
	err := UserLogin(c, userID, password)
	if err != nil {
		c.Echo().Logger.Debugf("User[%s] Login Error. [%s]", userID, err)
		msg := "The user ID or password is incorrect."
		data := map[string]string{"user_id": userID, "password": "", "msg": msg}
		return c.Render(http.StatusOK, "login", data)
	}
	isAdmin, err := CheckRoleByUserID(userID, model.RoleAdmin)
	if err != nil {
		c.Echo().Logger.Debugf("Admin Role Check Error. [%s]", userID, err)
		isAdmin = false
	}
	if isAdmin {
		c.Echo().Logger.Debugf("User is Admin. [%s]", userID)
		return c.Redirect(http.StatusTemporaryRedirect, "/admin")
	}
	return c.Redirect(http.StatusTemporaryRedirect, "/users/"+userID)
}

func handleLogoutPost(c echo.Context) error {
	err := UserLogout(c)
	if err != nil {
		c.Echo().Logger.Debugf("User Logout Error. [%s]", err)
		return c.Render(http.StatusOK, "login", nil)
	}
	msg := "Logout."
	data := map[string]string{"user_id": "", "password": "", "msg": msg}
	return c.Render(http.StatusOK, "login", data)
}
