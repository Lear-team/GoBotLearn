package sqlapi

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	apitypes "GoBotPigeon/types/apitypes"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type API struct {
	db *sqlx.DB
}

func NewAPI(db *sqlx.DB) *API {
	return &API{
		db: db,
	}
}

// GetUserByID ...
func (api *API) GetUserByID(idUser string) (*apitypes.UserRow, error) {
	userRow := []apitypes.UserRow{}

	err := api.db.Select(&userRow, "SELECT * FROM prj_user WHERE (userid = $1)", idUser)
	if err != nil {
		log.Println("GetUserByID api.db.Select failed with an error: ", err.Error())
		return nil, err
	}

	if len(userRow) == 1 {
		return &userRow[0], err
	}

	return nil, err
}

// GetUserByName ...
func (api *API) GetUserByName(nameUser string) (*apitypes.UserRow, error) {
	userRow := apitypes.UserRow{}

	if err := api.db.Get(&userRow, "SELECT * FROM prj_user WHERE nameuser = $1 LIMIT 1;", nameUser); err != nil {
		log.Println("GetUserByName api.db.Select failed with an error: ", err.Error())
		return nil, err
	}

	return &userRow, nil
}

// GetCodeByID ...
func (api *API) GetCodeByID(idCode string) (*apitypes.CodeRow, error) {
	const query = `SELECT * 
		FROM prj_code 
		WHERE codeid = $1 
		LIMIT 1;
	`

	codeRow := apitypes.CodeRow{}
	if err := api.db.Get(&codeRow, query, idCode); err != nil {
		return nil, errors.Wrapf(err, "can't get code by id %s", idCode)
	}

	return &codeRow, nil
}

// GetCodeByCode ...
func (api *API) GetCodeByCode(codeCode string) (*apitypes.CodeRow, error) {
	return getCodeByCode(context.Background(), api.db, codeCode)
}

func getCodeByCode(ctx context.Context, db TxContext, codeCode string) (*apitypes.CodeRow, error) {
	codeRow := []apitypes.CodeRow{}

	err := db.SelectContext(ctx, &codeRow, "SELECT * FROM prj_code WHERE (code = $1)", codeCode)
	if err != nil {
		log.Println("GetCodeByCode api.db.Select failed with an error: ", err.Error())
		return nil, err
	}

	if len(codeRow) == 1 {
		return &codeRow[0], err
	}

	return nil, err
}

func (api *API) AddNewUser2(username string) error {
	fmt.Println(username)
	return nil
}

// AddNewUser ...
func (api *API) AddNewUser(userN, userID, chatID string) (*apitypes.UserRow, error) {
	user, err := api.GetUserByID(userID)

	if err != nil {
		log.Println("AddNewUser GetUserByID failed with an error: ", err.Error())
		return nil, err
	}

	if user != nil {
		return nil, err
	}

	tx := api.db.MustBegin()
	tx.MustExec(`INSERT INTO prj_user ("userid", "nameuser", "chatid") VALUES ($1, $2, $3)`,
		userID, userN, chatID)
	tx.Commit()

	user, err = api.GetUserByID(userID)
	if err != nil {
		log.Println("AddNewUser GetUserByID failed with an error: ", err.Error())
		return nil, err
	}

	return user, err
}

// AddNewCode ...
func (api *API) AddNewCode(codeN string) (*apitypes.CodeRow, error) {
	var code *apitypes.CodeRow
	var err error

	work := func(ctx context.Context, db TxContext) error {
		code, err = getCodeByCode(context.Background(), api.db, codeN)
		if err != nil {
			log.Println("AddNewCode GetCodeByCode failed with an error: ", err.Error())
			return err
		}

		if code != nil {
			return err
		}

		uuidWithHyphen := uuid.New()
		uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

		if _, err := db.ExecContext(ctx, `INSERT INTO prj_code ("codeid", "code") VALUES ($1, $2)`, uuid, codeN); err != nil {
			return err
		}

		// TODO: refactore it!
		code, err = api.GetCodeByID(uuid)
		if err != nil {
			log.Println("AddNewCode GetCodeByID failed with an error: ", err.Error())
			return err
		}

		return nil
	}

	if err := RunInTransaction(context.Background(), api.db, work); err != nil {
		return nil, err
	}

	return code, nil
}

// AddRefUserCode ...
func (api *API) AddRefUserCode(codeR string, userIDR string) (*apitypes.RefUserCode, error) {
	user, err := api.GetUserByID(userIDR)

	if err != nil {
		// log.Fatal(err)
		return nil, err
	}

	if user == nil {
		log.Printf("Такой пользователя не зарегестрирован")
		return nil, err
	}

	refUserCode, err := api.GetRefUserCodeByUserName(user.NameUser)
	if err != nil {
		log.Println("AddRefUserCode GetRefUserCodeByUserName failed with an error: ", err.Error())
		return nil, err
	}

	if refUserCode != nil {
		log.Printf("Пользователь уже установил кодовое слово")
		return nil, err
	}

	code, err := api.GetCodeByCode(codeR)
	if err != nil {
		log.Println("AddRefUserCode GetCodeByCode failed with an error: ", err.Error())
		return nil, err
	}

	if code == nil {
		code, err = api.AddNewCode(codeR)
	}

	if err != nil {
		log.Println("AddRefUserCode AddNewCode failed with an error: ", err.Error())
		return nil, err
	}

	if code == nil {
		//log.Fatal(err)
		return nil, err
	}

	uuidWithHyphen := uuid.New()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

	tx := api.db.MustBegin()
	tx.MustExec(`INSERT INTO ref_usercode ("keyid", "codeid", "userid") VALUES ($1, $2, $3)`,
		uuid, code.CodeID, user.UserID)
	tx.Commit()

	refUserCode, err = api.GetRefUserCodeByKeyID(uuid)
	if err != nil {
		log.Println("AddRefUserCode GetRefUserCodeByKeyID failed with an error: ", err.Error())
		return nil, err
	}

	return refUserCode, err
}

// GetRefUserCodeByKeyID ...
func (api *API) GetRefUserCodeByKeyID(keyID string) (*apitypes.RefUserCode, error) {
	refUserCode := []apitypes.RefUserCode{}
	err := api.db.Select(&refUserCode, "SELECT * FROM ref_usercode WHERE (keyid = $1)", keyID)

	if err != nil {
		log.Println("GetRefUserCodeByKeyID api.db.Select failed with an error: ", err.Error())
		return nil, err
	}

	if len(refUserCode) == 1 {
		return &refUserCode[0], err
	}

	return nil, err
}

// GetRefUserCodeByUserName ...
func (api *API) GetRefUserCodeByUserName(userN string) (*apitypes.RefUserCode, error) {
	user, err := api.GetUserByName(userN)

	if err != nil {
		log.Println("GetRefUserCodeByUserName GetUserByName failed with an error: ", err.Error())
		return nil, err
	}

	if user == nil {
		log.Printf("Такой пользователя не зарегестрирован")
		return nil, err
	}

	rowUserCode := []apitypes.RefUserCode{}
	err = api.db.Select(&rowUserCode, "SELECT * FROM ref_usercode WHERE (userid = $1)", string(user.UserID))

	if err != nil {
		log.Println("GetRefUserCodeByUserName api.db.Select failed with an error: ", err.Error())
		return nil, err
	}

	if len(rowUserCode) == 1 {
		return &rowUserCode[0], nil
	}

	return nil, err
}

// UpdateRefUserCode ...
func (api *API) UpdateRefUserCode(codeR string, userR string) (*apitypes.RefUserCode, error) {

	user, err := api.GetUserByName(userR)

	if err != nil {
		log.Println("UpdateRefUserCode GetUserByName failed with an error: ", err.Error())
		return nil, err
	}

	if user == nil {
		log.Printf("Такой пользователя не зарегестрирован")
		return nil, err
	}

	refUserCode, err := api.GetRefUserCodeByUserName(user.NameUser)
	if err != nil {
		log.Println("UpdateRefUserCode GetRefUserCodeByUserName failed with an error: ", err.Error())
		return nil, err
	}

	code, err := api.GetCodeByCode(codeR)
	if err != nil {
		log.Println("UpdateRefUserCode GetCodeByCode failed with an error: ", err.Error())
		return nil, err
	}

	if code == nil {
		code, err = api.AddNewCode(codeR)
	}

	if err != nil {
		log.Println("UpdateRefUserCode AddNewCode failed with an error: ", err.Error())
		return nil, err
	}

	if refUserCode == nil {
		//log.Fatal("Пользователь еще не установил кодовое слово") // можно сразу вызывать добавление, но думаю лучше обработать ошибку
		refUserCode, err = api.AddRefUserCode(codeR, userR)

		if err != nil {
			log.Println("UpdateRefUserCode AddRefUserCode failed with an error: ", err.Error())
			return nil, err
		}
		return refUserCode, err
	}

	tx := api.db.MustBegin()
	tx.MustExec(`UPDATE ref_usercode SET codeid = $1 WHERE keyid = $2`, code.CodeID, refUserCode.KeyID)
	tx.Commit()

	refUserCode, err = api.GetRefUserCodeByKeyID(refUserCode.KeyID)
	if err != nil {
		log.Println("UpdateRefUserCode GetRefUserCodeByKeyID failed with an error: ", err.Error())
		return nil, err
	}

	return refUserCode, err
}

// CheckingPigeonWork ...
func (api *API) CheckingPigeonWork(userN string) (bool, error) {
	user, err := api.GetUserByName(userN)

	if err != nil {
		log.Println("CheckingPigeonWork GetUserByName failed with an error: ", err.Error())
		return false, err
	}

	var botWork = []apitypes.BotWork{}
	err = api.db.Select(&botWork, "SELECT * FROM prj_botwork WHERE (userid = $1)", user.UserID)

	if err != nil {
		log.Println("CheckingPigeonWork api.db.Select failed with an error: ", err.Error())
		return false, err
	}

	if len(botWork) == 1 {
		return botWork[0].BotWorkFlag, err
	}

	if len(botWork) > 1 {
		log.Printf("Ошибка с флагом работы бота")
		return false, err
	}

	return false, err
}

// StartPigeonWork ...
func (api *API) StartPigeonWork(userN string) error {
	user, err := api.GetUserByName(userN)

	if err != nil {
		log.Println("StartPigeonWork GetUserByName failed with an error: ", err.Error())
		return err
	}

	var botWork = []apitypes.BotWork{}
	err = api.db.Select(&botWork, "SELECT * FROM prj_botwork WHERE (userid = $1)", user.UserID)

	if err != nil {
		log.Println("StartPigeonWork api.db.Select failed with an error: ", err.Error())
		return err
	}

	if len(botWork) == 1 {
		tx := api.db.MustBegin()
		tx.MustExec("UPDATE prj_botwork SET botworkflag = $1  WHERE botworkid = $2 ", true, botWork[0].BotWorkID)
		tx.Commit()
	}

	if len(botWork) == 0 {
		err = api.CreatePigeonWorkFlag(userN)
		if err != nil {
			log.Println("StartPigeonWork CreatePigeonWorkFlag failed with an error: ", err.Error())
			return err
		}
		err = api.db.Select(&botWork, "SELECT * FROM prj_botwork WHERE (userid = $1)", user.UserID)
		if err != nil {
			log.Println("StartPigeonWork api.db.Select failed with an error: ", err.Error())
			return err
		}
	}

	return err
}

// StopPigeonWork ...
func (api *API) StopPigeonWork(userN string) error {
	user, err := api.GetUserByName(userN)

	if err != nil {
		log.Println("StopPigeonWork GetUserByName failed with an error: ", err.Error())
		return err
	}

	var botWork = []apitypes.BotWork{}
	err = api.db.Select(&botWork, "SELECT * FROM prj_botwork WHERE userid = $1", user.UserID)

	if err != nil {
		log.Println("StopPigeonWork api.db.Select failed with an error: ", err.Error())
		return err
	}

	if len(botWork) == 1 {
		tx := api.db.MustBegin()
		tx.MustExec("UPDATE prj_botwork SET botworkflag = $1 WHERE botworkid = $2", false, botWork[0].BotWorkID)
		tx.Commit()
	}

	if len(botWork) > 1 {
		log.Printf("Больше двух флагов")
		return err
	}

	return err
}

// CreatePigeonWorkFlag ...
func (api *API) CreatePigeonWorkFlag(userN string) error {
	user, err := api.GetUserByName(userN)

	if err != nil {
		log.Println("CreatePigeonWorkFlag GetUserByName failed with an error: ", err.Error())
		return err
	}

	var botWork = []apitypes.BotWork{}
	err = api.db.Select(&botWork, "SELECT * FROM prj_botwork WHERE userid = $1", user.UserID)

	if err != nil {
		log.Println("CreatePigeonWorkFlag api.db.Select failed with an error: ", err.Error())
		return err
	}

	if len(botWork) == 0 {

		uuidWithHyphen := uuid.New()
		uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

		tx := api.db.MustBegin()
		tx.MustExec(`INSERT INTO prj_botwork ("botworkid", "userid", "botworkflag") VALUES ($1, $2, $3)`,
			uuid, user.UserID, true)
		tx.Commit()

		return err
	}
	log.Printf("Флаг уже создан")
	return err
}

// SetLastComandUser ...
func (api *API) SetLastComandUser(userN string, command string) error {
	user, err := api.GetUserByName(userN)

	if err != nil {
		log.Println("SetLastComandUser GetUserByName failed with an error: ", err.Error())
		return err
	}

	uuidWithHyphen := uuid.New()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	today := time.Now()
	tTime := today.Add(10 * time.Minute).Format("2006/1/2 15:04")
	tx := api.db.MustBegin()

	tx.MustExec(`INSERT INTO prj_lastusercommand ("commandid", "userid", "command", "datacommand") VALUES ($1, $2, $3, $4)`,
		uuid, user.UserID, command, tTime)
	tx.Commit()

	return err
}

// GetLastCommandByUserName ...
func (api *API) GetLastCommandByUserName(userN string) (*apitypes.LastUserCommand, error) {
	user, err := api.GetUserByName(userN)

	if err != nil {
		log.Println("GetLastCommandByUserName GetUserByName failed with an error: ", err.Error())
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	var arrCommand = []apitypes.LastUserCommand{}

	err = api.db.Select(&arrCommand, "SELECT * FROM prj_lastusercommand WHERE (userid = $1) ORDER BY datacommand DESC", user.UserID)

	if err != nil {
		log.Println("GetLastCommandByUserName api.db.Select failed with an error: ", err.Error())
		return nil, err
	}

	if len(arrCommand) == 0 {
		return nil, nil
	}

	return &arrCommand[0], nil
}

// DeleteLastCommand ...
func (api *API) DeleteLastCommand(userN string, command string) error {
	user, err := api.GetUserByName(userN)

	if err != nil {
		log.Println("DeleteLastCommand GetUserByNamefailed with an error: ", err.Error())
		return err
	}
	tx := api.db.MustBegin()
	tx.MustExec("DELETE FROM prj_lastusercommand WHERE userid = $1", user.UserID)
	tx.Commit()

	return nil
}
