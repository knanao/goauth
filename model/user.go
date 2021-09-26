package model

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"

	"github.com/labstack/echo"
)

type (
	ID        string
	StringMD5 string
	Role      string
)

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	ID       ID        `json:"id"`
	UserID   string    `json:"user_id"`
	Password StringMD5 `json:"password"`
	FullName string    `json:"full_name"`
	Roles    []Role    `json:"roles"`
}

func (u *User) Copy(f *User) {
	u.ID = f.ID
	u.UserID = f.UserID
	u.Password = f.Password
	u.FullName = f.FullName
	u.Roles = make([]Role, len(f.Roles))
	copy(u.Roles, f.Roles)
}

type UserDataAccessor struct {
	stopCh    chan struct{}
	commandCh chan command
}

func (a *UserDataAccessor) Start(echo *echo.Echo) error {
	e = echo
	users = make(map[ID]User)
	if err := a.decodeJSON(); err != nil {
		return err
	}
	go a.mainLoop()
	return nil
}

func (a *UserDataAccessor) Stop() {
	a.stopCh <- struct{}{}
}

func (a *UserDataAccessor) FindAll() ([]User, error) {
	respCh := make(chan response, 1)
	defer close(respCh)
	req := []interface{}{}
	cmd := command{commandFindAll, req, respCh}
	a.commandCh <- cmd
	resp := <-respCh
	var res []User
	if resp.err != nil {
		e.Logger.Debugf("User Find Error. [%s]", resp.err)
		return res, resp.err
	}
	if res, ok := resp.result[0].([]User); ok {
		return res, nil
	}
	e.Logger.Debugf("User Find Error. [%s]", ErrorOther)
	return res, ErrorOther
}

func (a *UserDataAccessor) FindByUserID(reqUserID string, option FindOption) ([]User, error) {
	respCh := make(chan response, 1)
	defer close(respCh)
	req := []interface{}{reqUserID, option}
	cmd := command{commandFindByUserID, req, respCh}
	a.commandCh <- cmd
	resp := <-respCh
	var res []User
	if resp.err != nil {
		e.Logger.Debugf("User[UserID=%s] Find Error. [%s]", reqUserID, resp.err)
		return res, resp.err
	}
	if res, ok := resp.result[0].([]User); ok {
		return res, nil
	}
	e.Logger.Debugf("User[UserID=%s] Find Error. [%s]", reqUserID, ErrorOther)
	return res, ErrorOther
}

func EncodeStringMD5(str string) StringMD5 {
	h := md5.New()
	io.WriteString(h, str)
	encodeStr := hex.EncodeToString(h.Sum(nil))
	res := StringMD5(encodeStr)

	return res
}

type FindOption int

const (
	FIndAll    FindOption = iota // 全件検索
	FindFirst                    // 1件目のみ返す
	FindUnique                   // 結果が1件のみでない場合にはエラーを返す
)

var (
	ErrorNotFound        = errors.New("Not found")
	ErrorMultipleResults = errors.New("Multiple results")
	ErrorInvalidCommand  = errors.New("Invalid Command")
	ErrorBadParameter    = errors.New("Bad Parameter")
	ErrorNotImplemented  = errors.New("Not Implemented")
	ErrorOther           = errors.New("Other")
)

func (a *UserDataAccessor) decodeJSON() error {
	bytes, err := ioutil.ReadFile("../data/users.json")
	if err != nil {
		return err
	}
	var records []User
	if err := json.Unmarshal(bytes, &records); err != nil {
		return err
	}
	for _, x := range records {
		users[x.ID] = x
	}
	return nil
}

var e *echo.Echo

var users map[ID]User

type commandType int

const (
	commandFindAll      commandType = iota // 全件検索
	commandFindByID                        // IDで検索
	commandFindByUserID                    // UserIDで検索
)

type command struct {
	cmdType    commandType
	req        []interface{}
	responseCh chan response
}

type response struct {
	result []interface{}
	err    error
}

func (a *UserDataAccessor) mainLoop() {
	a.stopCh = make(chan struct{}, 1)
	a.commandCh = make(chan command, 1)
	defer close(a.commandCh)
	defer close(a.stopCh)
	e.Logger.Info("model.UserDataAccessor:start")
loop:
	for {
		select {
		case cmd := <-a.commandCh:
			switch cmd.cmdType {
			case commandFindAll:
				results := []User{}
				for _, x := range users {
					user := User{}
					user.Copy(&x)
					results = append(results, user)
				}
				res := []interface{}{results}
				cmd.responseCh <- response{res, nil}
				break
			case commandFindByID:
				cmd.responseCh <- response{nil, ErrorNotImplemented}
				break
			case commandFindByUserID:
				reqUserID, ok := cmd.req[0].(string)
				if !ok {
					cmd.responseCh <- response{nil, ErrorBadParameter}
					break
				}
				reqOption, ok := cmd.req[1].(FindOption)
				if !ok {
					cmd.responseCh <- response{nil, ErrorBadParameter}
					break
				}
				results := []User{}
				for _, x := range users {
					if x.UserID == reqUserID {
						user := User{}
						user.Copy(&x)
						results = append(results, user)
						if reqOption == FindFirst {
							break
						}
					}
				}
				if len(results) <= 0 {
					cmd.responseCh <- response{nil, ErrorNotFound}
					break
				}
				if reqOption == FindUnique && len(results) > 1 {
					cmd.responseCh <- response{nil, ErrorMultipleResults}
					break
				}
				res := []interface{}{results}
				cmd.responseCh <- response{res, nil}
			default:
				cmd.responseCh <- response{nil, ErrorInvalidCommand}
			}
		case <-a.stopCh:
			break loop
		}
	}
	e.Logger.Info("model.UserDataAccessor:stop")
}
