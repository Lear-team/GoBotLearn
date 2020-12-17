package sqlapi

import (
	"log"
	"strings"

	apitypes "GoBotPigeon/types/apitypes"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // here
)

// DB ...
var DB *sqlx.DB

// ConnectDb ...
func ConnectDb(databaseURL string) (err error) {

	DB, err = sqlx.Open("postgres", databaseURL)
	if err != nil {
		// log.Fatal(err)
		return err
	}

	if err := DB.Ping(); err != nil {
		// log.Fatal(err)
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
		// log.Fatal(err)
		return nil, err
	}

	err := db.Select(&userRow, "SELECT * FROM prj_user WHERE (userid = $1)", idUser)

	if err != nil {
		//log.Fatal(err)
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
		// log.Fatal(err)
		return nil, err
	}

	err := db.Select(&userRow, "SELECT * FROM prj_user WHERE (nameuser = $1)", nameUser)

	if err != nil {
		//log.Fatal(err)
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
		// log.Fatal(err)
		return nil, err
	}

	err := db.Select(&codeRow, "SELECT * FROM prj_code WHERE (codeid = $1)", idCode)

	if err != nil {
		// log.Fatal(err)
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
		// log.Fatal(err)
		return nil, err
	}

	err := db.Select(&codeRow, "SELECT * FROM prj_code WHERE (code = $1)", codeCode)

	if err != nil {
		//log.Fatal(err)
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
		//log.Fatal(err)
		return nil, err
	}

	// user, err := GetUserByName(userN, db)
	user, err := GetUserByID(userID, db)

	if err != nil {
		//log.Fatal(err)
		return nil, err
	}

	if user != nil {
		return nil, err
	}

	// uuidWithHyphen := uuid.New()
	// uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

	tx := db.MustBegin()
	tx.MustExec(`INSERT INTO prj_user ("userid", "nameuser", "chatid") VALUES ($1, $2, $3)`,
		userID, userN, chatID)
	tx.Commit()

	user, err = GetUserByID(userID, db)
	if err != nil {
		//log.Fatal(err)
		return nil, err
	}

	return user, err
}

// AddNewCode ...
func AddNewCode(codeN string, db *sqlx.DB) (*apitypes.CodeRow, error) {

	if err := db.Ping(); err != nil {
		// log.Fatal(err)
		return nil, err
	}

	code, err := GetCodeByCode(codeN, db)
	if err != nil {
		// log.Fatal(err)
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
		// log.Fatal(err)
		return nil, err
	}

	return code, err
}

// AddRefUserCode ...
func AddRefUserCode(codeR string, userIDR string, db *sqlx.DB) (*apitypes.RefUserCode, error) {
	// var user apitypes.UserRow
	var refUserCode = &apitypes.RefUserCode{}

	if err := db.Ping(); err != nil {
		// log.Fatal(err)
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
		// log.Fatal(err)
		return nil, err
	}

	if refUserCode != nil {
		log.Printf("Пользователь уже установил кодовое слово")
		return nil, err
	}

	code, err := GetCodeByCode(codeR, db)
	if err != nil {
		//log.Fatal(err)
		return nil, err
	}

	if code == nil {
		code, err = AddNewCode(codeR, db)
	}

	if err != nil {
		//log.Fatal(err)
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
		// log.Fatal(err)
		return nil, err
	}

	return refUserCode, err
}

// GetRefUserCodeByKeyID ...
func GetRefUserCodeByKeyID(keyID string, db *sqlx.DB) (*apitypes.RefUserCode, error) {

	if err := db.Ping(); err != nil {
		//log.Fatal(err)
		return nil, err
	}

	refUserCode := []apitypes.RefUserCode{}
	err := db.Select(&refUserCode, "SELECT * FROM ref_usercode WHERE (keyid = $1)", keyID)

	if err != nil {
		//log.Fatal(err)
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
		// log.Fatal(err)
		return nil, err
	}

	user, err := GetUserByName(userN, db)

	if err != nil {
		//log.Fatal(err)
		return nil, err
	}

	if user == nil {
		log.Printf("Такой пользователя не зарегестрирован")
		return nil, err
	}

	rowUserCode := []apitypes.RefUserCode{}
	err = db.Select(&rowUserCode, "SELECT * FROM ref_usercode WHERE (userid = $1)", string(user.UserID))

	if err != nil {
		//log.Fatal(err)
		return nil, err
	}

	if len(rowUserCode) == 1 {
		return &rowUserCode[0], nil
	}

	return nil, err
}

// CheckRefUserCode ...
func CheckRefUserCode(codeID string, userID string, db *sqlx.DB) (bool, error) {

	if err := db.Ping(); err != nil {
		// log.Fatal(err)
		return false, err
	}

	refUserCode := []apitypes.RefUserCode{}
	err := db.Select(&refUserCode, "SELECT * FROM ref_usercode WHERE (userid = $1 AND codeid = $2)", userID, codeID)

	if err != nil {
		// log.Fatal(err)
		return false, err
	}

	if len(refUserCode) == 1 {
		return true, err
	}

	if len(refUserCode) > 1 {
		log.Printf("Ошибка в связке Пользователь-Кодовое слово")
		return false, err
	}

	return false, err
}

// DeleteRefUserCodeByUserName ...
func DeleteRefUserCodeByUserName(userN string, db *sqlx.DB) (*apitypes.RefUserCode, error) {

	var refUserCode = &apitypes.RefUserCode{}

	if err := db.Ping(); err != nil {
		// log.Fatal(err)
		return nil, err
	}

	refUserCode, err := GetRefUserCodeByUserName(userN, db)

	if err != nil {
		//log.Fatal(err)
		return nil, err
	}

	tx := db.MustBegin()
	tx.MustExec(`UPDATE ref_usercode codeid = "" WHERE keyid = '$1'`, refUserCode.KeyID)
	tx.Commit()

	refUserCode, err = GetRefUserCodeByUserName(userN, db)

	if refUserCode != nil {
		log.Printf("Удалить связь с кодовым словом не удалось")
		return nil, err
	}

	return refUserCode, err
}

// UpdateRefUserCode ...
func UpdateRefUserCode(codeR string, userR string, db *sqlx.DB) (*apitypes.RefUserCode, error) {
	var refUserCode = &apitypes.RefUserCode{}

	if err := db.Ping(); err != nil {
		//log.Fatal(err)
		return nil, err
	}

	user, err := GetUserByName(userR, db)

	if err != nil {
		//log.Fatal(err)
		return nil, err
	}

	if user == nil {
		log.Printf("Такой пользователя не зарегестрирован")
		return nil, err
	}

	refUserCode, err = GetRefUserCodeByUserName(user.NameUser, db)
	if err != nil {
		//log.Fatal(err)
		return nil, err
	}

	code, err := GetCodeByCode(codeR, db)
	if err != nil {
		//log.Fatal(err)
		return nil, err
	}

	if code == nil {
		code, err = AddNewCode(codeR, db)
	}

	if err != nil {
		//log.Fatal(err)
		return nil, err
	}

	if refUserCode == nil {
		//log.Fatal("Пользователь еще не установил кодовое слово") // можно сразу вызывать добавление, но думаю лучше обработать ошибку
		refUserCode, err = AddRefUserCode(codeR, userR, db)

		if err != nil {
			//log.Fatal(err)
			return nil, err
		}
		return refUserCode, err
	}

	tx := db.MustBegin()
	tx.MustExec(`UPDATE ref_usercode codeid = "$1" WHERE keyid = '$2'`,
		codeR, refUserCode.KeyID)
	tx.Commit()

	refUserCode, err = GetRefUserCodeByKeyID(refUserCode.KeyID, db)
	if err != nil {
		//log.Fatal(err)
		return nil, err
	}

	return refUserCode, err
}

//  протестировать API для работы с базой данных

//  проверить добавление связи пользоватль - кодовое слово
//  проверить обновление связи пользователь - кодовое слово

//  добавить функцию сохранения сообщения с сайта
//  добавить в тип сообщениня поле ДАТА  +