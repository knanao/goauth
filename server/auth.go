package main

import (
	"errors"
	"net/http"

	"github.com/knanao/goauth/server/model"
	"github.com/knanao/goauth/server/session"
	"github.com/labstack/echo"
)

var (
	ErrorInvalidUserID   = errors.New("Invalid UserID")
	ErrorInvalidPassword = errors.New("Invalid Password")
	ErrorNotLoggedIn     = errors.New("Not Logged In")
)

func UserLogin(c echo.Context, userID string, password string) error {
	users, err := userDA.FindByUserID(userID, model.FindFirst)
	if err != nil {
		return err
	}
	user := &users[0]
	encodePassword := model.EncodeStringMD5(password)
	if user.Password != encodePassword {
		return ErrorInvalidPassword
	}
	sessionID, err := sessionManager.Create()
	if err != nil {
		return err
	}
	err = session.WriteCookie(c, sessionID)
	if err != nil {
		return err
	}
	sessionStore, err := sessionManager.LoadStore(sessionID)
	if err != nil {
		return err
	}
	sessionData := map[string]string{
		"user_id": userID,
	}
	sessionStore.Data = sessionData
	err = sessionManager.SaveStore(sessionID, sessionStore)
	if err != nil {
		return err
	}

	return nil
}

func UserLogout(c echo.Context) error {
	sessionID, err := session.ReadCookie(c)
	if err != nil {
		return err
	}
	err = sessionManager.Delete(sessionID)
	if err != nil {
		return err
	}

	return nil
}

func CheckUserID(c echo.Context, userID string) error {
	sessionID, err := session.ReadCookie(c)
	if err != nil {
		return err
	}
	sessionStore, err := sessionManager.LoadStore(sessionID)
	if err != nil {
		return err
	}
	sessionUserID, ok := sessionStore.Data["user_id"]
	if !ok {
		return ErrorNotLoggedIn
	}
	if sessionUserID != userID {
		return ErrorInvalidUserID
	}

	return nil
}

func CheckRole(c echo.Context, role model.Role) (bool, error) {
	sessionID, err := session.ReadCookie(c)
	if err != nil {
		return false, err
	}
	sessionStore, err := sessionManager.LoadStore(sessionID)
	if err != nil {
		return false, err
	}
	sessionUserID, ok := sessionStore.Data["user_id"]
	if !ok {
		return false, ErrorNotLoggedIn
	}
	haveRole, err := CheckRoleByUserID(sessionUserID, role)
	return haveRole, nil
}

func CheckRoleByUserID(userID string, role model.Role) (bool, error) {
	users, err := userDA.FindByUserID(userID, model.FindFirst)
	if err != nil {
		return false, err
	}
	user := &users[0]
	for _, v := range user.Roles {
		if v == role {
			return true, nil
		}
	}

	return false, nil
}

func MiddlewareAuthAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		isAdmin, err := CheckRole(c, model.RoleAdmin)
		if err != nil {
			c.Echo().Logger.Debugf("Admin Page Role Error. [%s]", err)
			isAdmin = false
		}
		if !isAdmin {
			msg := "You have not logged in as an admin."
			return c.Render(http.StatusOK, "error", msg)
		}
		return next(c)
	}
}
