package sqlapi

import (
	"log"
	"strings"
	"time"

	apitypes "GoBotPigeon/types/apitypes"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // here
	"github.com/pkg/errors"
)

// DB ...
var DB *sqlx.DB

// ConnectDb ...
func ConnectDb(databaseURL string) (err error) {

	DB, err = sqlx.Open("postgres", databaseURL)
	if err != nil {
		log.Println("sqlx.Open failed with an error: ", err.Error())
		return err
	}

	if err := DB.Ping(); err != nil {
		log.Println("DB.Ping failed with an error: ", err.Error())
		return err
	}

	return
}

// CloseConnectDb ...
func CloseConnectDb(db *sqlx.DB) {
	db.Close()
}

// GetUserByID ...
func GetUserByID(idUser string, db *sqlx.DB) (*apitypes.UserRow, error) {
	userRow := []apitypes.UserRow{}

	if err := db.Ping(); err != nil {
		log.Println("DB.Ping failed with an error: ", err.Error())
		return nil, err
	}

	err := db.Select(&userRow, "SELECT * FROM prj_user WHERE (userid = $1)", idUser)

	if err != nil {
		log.Println("GetUserByID db.Select failed with an error: ", err.Error())
		return nil, err
	}

	if len(userRow) == 1 {
		return &userRow[0], err
	}

	return nil, err
}

// GetUserByName ...
func GetUserByName(nameUser string, db *sqlx.DB) (*apitypes.UserRow, error) {

	userRow := []apitypes.UserRow{}
	if err := db.Ping(); err != nil {
		log.Println("GetUserByName db.Ping failed with an error: ", err.Error())
		return nil, err
	}

	err := db.Select(&userRow, "SELECT * FROM prj_user WHERE (nameuser = $1)", nameUser)

	if err != nil {
		log.Println("GetUserByName db.Select failed with an error: ", err.Error())
		return nil, err
	}

	if len(userRow) == 1 {
		return &userRow[0], err
	}

	return nil, err
}

// GetCodeByID ...
func GetCodeByID(idCode string, db *sqlx.DB) (*apitypes.CodeRow, error) {

	codeRow := []apitypes.CodeRow{}
	if err := db.Ping(); err != nil {
		log.Println("GetCodeByID db.Ping failed with an error: ", err.Error())
		return nil, err
	}

	err := db.Select(&codeRow, "SELECT * FROM prj_code WHERE (codeid = $1)", idCode)

	if err != nil {
		log.Println("GetCodeByID db.Select failed with an error: ", err.Error())
		return nil, err
	}

	if len(codeRow) == 1 {
		return &codeRow[0], err
	}

	return nil, err
}

// GetCodeByCode ...
func GetCodeByCode(codeCode string, db *sqlx.DB) (*apitypes.CodeRow, error) {

	codeRow := []apitypes.CodeRow{}
	if err := db.Ping(); err != nil {
		log.Println("GetCodeByCode db.Ping failed with an error: ", err.Error())
		return nil, err
	}

	err := db.Select(&codeRow, "SELECT * FROM prj_code WHERE (code = $1)", codeCode)

	if err != nil {
		log.Println("GetCodeByCode db.Select failed with an error: ", err.Error())
		return nil, err
	}

	if len(codeRow) == 1 {
		return &codeRow[0], err
	}

	return nil, err
}

// AddNewUser ...
func AddNewUser(userN string, userID string, chatID string, db *sqlx.DB) (*apitypes.UserRow, error) {

	if err := db.Ping(); err != nil {
		log.Println("AddNewUser db.Ping failed with an error: ", err.Error())
		return nil, err
	}

	user, err := GetUserByID(userID, db)

	if err != nil {
		log.Println("AddNewUser GetUserByID failed with an error: ", err.Error())
		return nil, err
	}

	if user != nil {
		return nil, err
	}

	tx := db.MustBegin()
	tx.MustExec(`INSERT INTO prj_user ("userid", "nameuser", "chatid") VALUES ($1, $2, $3)`,
		userID, userN, chatID)
	tx.Commit()

	user, err = GetUserByID(userID, db)
	if err != nil {
		log.Println("AddNewUser GetUserByID failed with an error: ", err.Error())
		return nil, err
	}

	return user, err
}

// AddNewCode ...
func AddNewCode(codeN string, db *sqlx.DB) (*apitypes.CodeRow, error) {

	if err := db.Ping(); err != nil {
		log.Println("AddNewCode db.Ping failed with an error: ", err.Error())
		return nil, err
	}

	code, err := GetCodeByCode(codeN, db)
	if err != nil {
		log.Println("AddNewCode GetCodeByCode failed with an error: ", err.Error())
		return nil, err
	}

	if code != nil {
		return nil, err
	}

	uuidWithHyphen := uuid.New()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

	tx := db.MustBegin()
	tx.MustExec(`INSERT INTO prj_code ("codeid", "code") VALUES ($1, $2)`,
		uuid, codeN)
	tx.Commit()

	code, err = GetCodeByID(uuid, db)
	if err != nil {
		log.Println("AddNewCode GetCodeByID failed with an error: ", err.Error())
		return nil, err
	}

	return code, err
}

// AddRefUserCode ...
func AddRefUserCode(codeR string, userIDR string, db *sqlx.DB) (*apitypes.RefUserCode, error) {
	var refUserCode = &apitypes.RefUserCode{}

	if err := db.Ping(); err != nil {
		log.Println("AddRefUserCode db.Ping failed with an error: ", err.Error())
		return nil, err
	}

	user, err := GetUserByID(userIDR, db)

	if err != nil {
		// log.Fatal(err)
		return nil, err
	}

	if user == nil {
		log.Printf("Такой пользователя не зарегестрирован")
		return nil, err
	}

	refUserCode, err = GetRefUserCodeByUserName(user.NameUser, db)
	if err != nil {
		log.Println("AddRefUserCode GetRefUserCodeByUserName failed with an error: ", err.Error())
		return nil, err
	}

	if refUserCode != nil {
		log.Printf("Пользователь уже установил кодовое слово")
		return nil, err
	}

	code, err := GetCodeByCode(codeR, db)
	if err != nil {
		log.Println("AddRefUserCode GetCodeByCode failed with an error: ", err.Error())
		return nil, err
	}

	if code == nil {
		code, err = AddNewCode(codeR, db)
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

	tx := db.MustBegin()
	tx.MustExec(`INSERT INTO ref_usercode ("keyid", "codeid", "userid") VALUES ($1, $2, $3)`,
		uuid, code.CodeID, user.UserID)
	tx.Commit()

	refUserCode, err = GetRefUserCodeByKeyID(uuid, db)
	if err != nil {
		log.Println("AddRefUserCode GetRefUserCodeByKeyID failed with an error: ", err.Error())
		return nil, err
	}

	return refUserCode, err
}

// GetRefUserCodeByKeyID ...
func GetRefUserCodeByKeyID(keyID string, db *sqlx.DB) (*apitypes.RefUserCode, error) {

	if err := db.Ping(); err != nil {
		log.Println("GetRefUserCodeByKeyID db.Ping failed with an error: ", err.Error())
		return nil, err
	}

	refUserCode := []apitypes.RefUserCode{}
	err := db.Select(&refUserCode, "SELECT * FROM ref_usercode WHERE (keyid = $1)", keyID)

	if err != nil {
		log.Println("GetRefUserCodeByKeyID db.Select failed with an error: ", err.Error())
		return nil, err
	}

	if len(refUserCode) == 1 {
		return &refUserCode[0], err
	}

	return nil, err
}

// GetRefUserCodeByUserName ...
func GetRefUserCodeByUserName(userN string, db *sqlx.DB) (*apitypes.RefUserCode, error) {

	if err := db.Ping(); err != nil {
		log.Println("GetRefUserCodeByUserName db.Ping failed with an error: ", err.Error())
		return nil, err
	}

	user, err := GetUserByName(userN, db)

	if err != nil {
		log.Println("GetRefUserCodeByUserName GetUserByName failed with an error: ", err.Error())
		return nil, err
	}

	if user == nil {
		log.Printf("Такой пользователя не зарегестрирован")
		return nil, err
	}

	rowUserCode := []apitypes.RefUserCode{}
	err = db.Select(&rowUserCode, "SELECT * FROM ref_usercode WHERE (userid = $1)", string(user.UserID))

	if err != nil {
		log.Println("GetRefUserCodeByUserName db.Select failed with an error: ", err.Error())
		return nil, err
	}

	if len(rowUserCode) == 1 {
		return &rowUserCode[0], nil
	}

	return nil, err
}

// UpdateRefUserCode ...
func UpdateRefUserCode(codeR string, userR string, db *sqlx.DB) (*apitypes.RefUserCode, error) {
	var refUserCode = &apitypes.RefUserCode{}

	if err := db.Ping(); err != nil {
		log.Println("UpdateRefUserCode db.Ping failed with an error: ", err.Error())
		return nil, err
	}

	user, err := GetUserByName(userR, db)

	if err != nil {
		log.Println("UpdateRefUserCode GetUserByName failed with an error: ", err.Error())
		return nil, err
	}

	if user == nil {
		log.Printf("Такой пользователя не зарегестрирован")
		return nil, err
	}

	refUserCode, err = GetRefUserCodeByUserName(user.NameUser, db)
	if err != nil {
		log.Println("UpdateRefUserCode GetRefUserCodeByUserName failed with an error: ", err.Error())
		return nil, err
	}

	code, err := GetCodeByCode(codeR, db)
	if err != nil {
		log.Println("UpdateRefUserCode GetCodeByCode failed with an error: ", err.Error())
		return nil, err
	}

	if code == nil {
		code, err = AddNewCode(codeR, db)
	}

	if err != nil {
		log.Println("UpdateRefUserCode AddNewCode failed with an error: ", err.Error())
		return nil, err
	}

	if refUserCode == nil {
		//log.Fatal("Пользователь еще не установил кодовое слово") // можно сразу вызывать добавление, но думаю лучше обработать ошибку
		refUserCode, err = AddRefUserCode(codeR, userR, db)

		if err != nil {
			log.Println("UpdateRefUserCode AddRefUserCode failed with an error: ", err.Error())
			return nil, err
		}
		return refUserCode, err
	}

	tx := db.MustBegin()
	tx.MustExec(`UPDATE ref_usercode SET codeid = $1 WHERE keyid = $2`, code.CodeID, refUserCode.KeyID)
	tx.Commit()

	refUserCode, err = GetRefUserCodeByKeyID(refUserCode.KeyID, db)
	if err != nil {
		log.Println("UpdateRefUserCode GetRefUserCodeByKeyID failed with an error: ", err.Error())
		return nil, err
	}

	return refUserCode, err
}

// CheckingPigeonWork ...
func CheckingPigeonWork(userN string, db *sqlx.DB) (bool, error) {

	if err := db.Ping(); err != nil {
		log.Println("CheckingPigeonWork db.Ping failed with an error: ", err.Error())
		return false, err
	}

	user, err := GetUserByName(userN, db)

	if err != nil {
		log.Println("CheckingPigeonWork GetUserByName failed with an error: ", err.Error())
		return false, err
	}

	var botWork = []apitypes.BotWork{}
	err = db.Select(&botWork, "SELECT * FROM prj_botwork WHERE (userid = $1)", user.UserID)

	if err != nil {
		log.Println("CheckingPigeonWork db.Select failed with an error: ", err.Error())
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
func StartPigeonWork(userN string, db *sqlx.DB) error {
	if err := db.Ping(); err != nil {
		log.Println("StartPigeonWork db.Ping failed with an error: ", err.Error())
		return err
	}

	user, err := GetUserByName(userN, db)

	if err != nil {
		log.Println("StartPigeonWork GetUserByName failed with an error: ", err.Error())
		return err
	}

	var botWork = []apitypes.BotWork{}
	err = db.Select(&botWork, "SELECT * FROM prj_botwork WHERE (userid = $1)", user.UserID)

	if err != nil {
		log.Println("StartPigeonWork db.Select failed with an error: ", err.Error())
		return err
	}

	if len(botWork) == 1 {
		tx := db.MustBegin()
		tx.MustExec("UPDATE prj_botwork SET botworkflag = $1  WHERE botworkid = $2 ", true, botWork[0].BotWorkID)
		tx.Commit()
	}

	if len(botWork) == 0 {
		err = CreatePigeonWorkFlag(userN, db)
		if err != nil {
			log.Println("StartPigeonWork CreatePigeonWorkFlag failed with an error: ", err.Error())
			return err
		}
		err = db.Select(&botWork, "SELECT * FROM prj_botwork WHERE (userid = $1)", user.UserID)
		if err != nil {
			log.Println("StartPigeonWork db.Select failed with an error: ", err.Error())
			return err
		}
	}

	return err
}

// StopPigeonWork ...
func StopPigeonWork(userN string, db *sqlx.DB) error {
	if err := db.Ping(); err != nil {
		log.Println("StopPigeonWork db.Ping failed with an error: ", err.Error())
		return err
	}

	user, err := GetUserByName(userN, db)

	if err != nil {
		log.Println("StopPigeonWork GetUserByName failed with an error: ", err.Error())
		return err
	}

	var botWork = []apitypes.BotWork{}
	err = db.Select(&botWork, "SELECT * FROM prj_botwork WHERE userid = $1", user.UserID)

	if err != nil {
		log.Println("StopPigeonWork db.Select failed with an error: ", err.Error())
		return err
	}

	if len(botWork) == 1 {
		tx := db.MustBegin()
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
func CreatePigeonWorkFlag(userN string, db *sqlx.DB) error {
	if err := db.Ping(); err != nil {
		log.Println("CreatePigeonWorkFlag db.Ping failed with an error: ", err.Error())
		return err
	}

	user, err := GetUserByName(userN, db)

	if err != nil {
		log.Println("CreatePigeonWorkFlag GetUserByName failed with an error: ", err.Error())
		return err
	}

	var botWork = []apitypes.BotWork{}
	err = db.Select(&botWork, "SELECT * FROM prj_botwork WHERE userid = $1", user.UserID)

	if err != nil {
		log.Println("CreatePigeonWorkFlag db.Select failed with an error: ", err.Error())
		return err
	}

	if len(botWork) == 0 {

		uuidWithHyphen := uuid.New()
		uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

		tx := db.MustBegin()
		tx.MustExec(`INSERT INTO prj_botwork ("botworkid", "userid", "botworkflag") VALUES ($1, $2, $3)`,
			uuid, user.UserID, true)
		tx.Commit()

		return err
	}
	log.Printf("Флаг уже создан")
	return err
}

// SetLastComandUser ...
func SetLastComandUser(userN string, db *sqlx.DB, command string) error {

	if err := db.Ping(); err != nil {
		log.Println("SetLastComandUser db.Ping failed with an error: ", err.Error())
		return err
	}

	user, err := GetUserByName(userN, db)

	if err != nil {
		log.Println("SetLastComandUser GetUserByName failed with an error: ", err.Error())
		return err
	}

	uuidWithHyphen := uuid.New()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	today := time.Now()
	tTime := today.Add(10 * time.Minute).Format("2006/1/2 15:04")
	tx := db.MustBegin()

	tx.MustExec(`INSERT INTO prj_lastusercommand ("commandid", "userid", "command", "datacommand") VALUES ($1, $2, $3, $4)`,
		uuid, user.UserID, command, tTime)
	tx.Commit()

	return err
}

// GetLastCommandByUserName ...
func GetLastCommandByUserName(userN string, db *sqlx.DB) (*apitypes.LastUserCommand, error) {

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "GetLastCommandByUserName db.Ping")
	}

	user, err := GetUserByName(userN, db)

	if err != nil {
		log.Println("GetLastCommandByUserName GetUserByName failed with an error: ", err.Error())
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	var arrCommand = []apitypes.LastUserCommand{}

	err = db.Select(&arrCommand, "SELECT * FROM prj_lastusercommand WHERE (userid = $1) ORDER BY datacommand DESC", user.UserID)

	if err != nil {
		log.Println("GetLastCommandByUserName db.Select failed with an error: ", err.Error())
		return nil, err
	}

	if len(arrCommand) == 0 {
		return nil, nil
	}

	return &arrCommand[0], nil
}

// DeleteLastCommand ...
func DeleteLastCommand(userN string, command string, db *sqlx.DB) error {
	if err := db.Ping(); err != nil {
		log.Println("DeleteLastCommand db.Ping failed with an error: ", err.Error())
		return err
	}

	user, err := GetUserByName(userN, db)

	if err != nil {
		log.Println("DeleteLastCommand GetUserByNamefailed with an error: ", err.Error())
		return err
	}
	tx := db.MustBegin()
	tx.MustExec("DELETE FROM prj_lastusercommand WHERE userid = $1", user.UserID)
	tx.Commit()

	return nil
}
