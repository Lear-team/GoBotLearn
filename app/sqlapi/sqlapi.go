package sqlapi

import (
	"context"
	"log"
	"strings"

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

// уйти от получения пользоваетля по имени, перейти на id
// GetUserByName ...
func (api *API) GetUserByName(ctx context.Context, nameUser string) (*apitypes.UserRow, error) {
	return getUserByName(ctx, api.db, nameUser)
}

func getUserByName(ctx context.Context, db TxContext, nameUser string) (*apitypes.UserRow, error) {
	userRow := []apitypes.UserRow{}

	query := "SELECT userid, nameuser, chatid FROM prj_user WHERE nameuser = $1 LIMIT 1;"

	err := db.SelectContext(ctx, &userRow, query, nameUser)

	if err != nil {
		return nil, errors.Wrapf(err, "can't get user by name %s", nameUser)
	}
	if len(userRow) == 0 {
		return nil, err
	}
	return &userRow[0], nil
}

// GetUserByID ...
func (api *API) GetUserByID(ctx context.Context, idUser string) (*apitypes.UserRow, error) {
	return getUserByID(ctx, api.db, idUser)
}

func getUserByID(ctx context.Context, db TxContext, userID string) (*apitypes.UserRow, error) {
	userRow := []apitypes.UserRow{}

	query := "SELECT userid, nameuser, chatid FROM prj_user WHERE userid = $1 LIMIT 1;"

	err := db.SelectContext(ctx, &userRow, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "getUserById failed")
	}

	if len(userRow) == 0 {
		return nil, err
	}

	return &userRow[0], err
}

// GetCodeByID ...
func (api *API) GetCodeByID(ctx context.Context, idCode string) (*apitypes.CodeRow, error) {
	return getCodeByID(ctx, api.db, idCode)
}

func getCodeByID(ctx context.Context, db TxContext, idCode string) (*apitypes.CodeRow, error) {

	const query = `SELECT codeid, code
		FROM prj_code 
		WHERE codeid = $1 
		LIMIT 1;
	`
	codeRow := []apitypes.CodeRow{}
	if err := db.SelectContext(ctx, &codeRow, query, idCode); err != nil {
		return nil, errors.Wrapf(err, "can't get code by id %s", idCode)
	}

	if len(codeRow) == 0 {
		return nil, nil
	}

	return &codeRow[0], nil
}

// GetCodeByCode ...
func (api *API) GetCodeByCode(ctx context.Context, codeCode string) (*apitypes.CodeRow, error) {
	return getCodeByCode(ctx, api.db, codeCode)
}

func getCodeByCode(ctx context.Context, db TxContext, codeCode string) (*apitypes.CodeRow, error) {
	codeRow := []apitypes.CodeRow{}

	query := "SELECT codeid, code FROM prj_code WHERE code = $1"

	err := db.SelectContext(ctx, &codeRow, query, codeCode)
	if err != nil {
		return nil, errors.Wrap(err, "GetCodeByCode api.db.Select failed with an error")
	}

	if len(codeRow) == 1 {
		return &codeRow[0], err
	}

	return nil, err
}

// GetRefUserCodeByKeyID ...
func (api *API) GetRefUserCodeByKeyID(ctx context.Context, keyID string) (*apitypes.RefUserCode, error) {
	return getRefUserCodeByKeyID(ctx, api.db, keyID)
}

func getRefUserCodeByKeyID(ctx context.Context, db TxContext, keyID string) (*apitypes.RefUserCode, error) {
	rowUserCode := []apitypes.RefUserCode{}

	query := "SELECT keyid, codeid, userid FROM ref_usercode WHERE keyid = $1 LIMIT 1;"

	err := db.SelectContext(ctx, &rowUserCode, query, keyID)
	if err != nil {
		return nil, errors.Wrapf(err, "can't get ref_usercode by keyid %s", keyID)
	}

	if len(rowUserCode) == 0 {
		return nil, nil
	}
	return &rowUserCode[0], nil
}

// GetRefUserCodeByUserName ...
func (api *API) GetRefUserCodeByUserName(ctx context.Context, userN string) (*apitypes.RefUserCode, error) {
	return getRefUserCodeByUserName(ctx, api.db, userN)
}

func getRefUserCodeByUserName(ctx context.Context, db TxContext, userN string) (*apitypes.RefUserCode, error) {

	query := `SELECT ref_usercode.keyid, ref_usercode.codeid, ref_usercode.userid FROM ref_usercode 
			JOIN prj_user ON ref_usercode.userid = prj_user.userid
			WHERE prj_user.nameuser = $1 LIMIT 1;`

	rowUserCode := []apitypes.RefUserCode{}
	err := db.SelectContext(ctx, &rowUserCode, query, userN)
	if err != nil {
		return nil, errors.Wrapf(err, "can't get ref_usercode by nameuser %s", userN)
	}

	if len(rowUserCode) == 0 {
		return nil, nil
	}

	return &rowUserCode[0], nil
}

// AddNewUser ...
func (api *API) AddNewUser(ctx context.Context, userN, userID, chatID string) (*apitypes.UserRow, error) {
	work := func(ctx context.Context, db TxContext) error {
		err := addNewUser(ctx, db, userN, userID, chatID)
		if err != nil {
			return err
		}

		return nil
	}

	if err := RunInTransaction(ctx, api.db, work); err != nil {
		return nil, err
	}

	var user *apitypes.UserRow
	user.ChatID = chatID
	user.UserID = userID
	user.NameUser = userN

	return user, nil
}

func addNewUser(ctx context.Context, db TxContext, userN, userID, chatID string) error {

	query := `INSERT INTO prj_user ("userid", "nameuser", "chatid") VALUES ($1, $2, $3)`

	if _, err := db.ExecContext(ctx, query,
		userID, userN, chatID); err != nil {
		return err
	}

	return nil
}

// AddNewCode ...
func (api *API) AddNewCode(ctx context.Context, codeN string) (*apitypes.CodeRow, error) {
	var code *apitypes.CodeRow
	var err error

	work := func(ctx context.Context, db TxContext) error {
		code, err = addNewCode(ctx, db, codeN)
		if err != nil {
			return err
		}
		return nil
	}

	if err := RunInTransaction(ctx, api.db, work); err != nil {
		return nil, err
	}

	return code, nil
}

func addNewCode(ctx context.Context, db TxContext, codeN string) (*apitypes.CodeRow, error) {
	uuidWithHyphen := uuid.New()
	uid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

	query := `INSERT INTO prj_code ("codeid", "code") VALUES ($1, $2)`

	if _, err := db.ExecContext(ctx, query, uid, codeN); err != nil {
		return nil, err
	}

	var code *apitypes.CodeRow
	code.Code = codeN
	code.CodeID = uid

	return code, nil
}

// AddRefUserCode ...
func (api *API) AddRefUserCode(ctx context.Context, codeR string, userIDR string) (*apitypes.RefUserCode, error) {
	var err error
	var uid string

	work := func(ctx context.Context, db TxContext) error {
		refUserCode := []apitypes.RefUserCode{} // перенес в функцию, но мне не нравится это решение, надо уточнить как правильно делать

		query := `SELECT ref_usercode.keyid, ref_usercode.codeid, ref_usercode.userid 
					FROM ref_usercode WHERE ref_usercode.userid = $1 LIMIT 1;`

		err := db.SelectContext(ctx, &refUserCode, query, userIDR)
		if err != nil {
			return errors.Wrap(err, "getRefUserCodeByUserName failed")
		}

		if len(refUserCode) != 0 {
			log.Printf("Пользователь уже установил кодовое слово")
			return err
		}

		code, err := getCodeByCode(ctx, api.db, codeR)
		if err != nil {
			return errors.Wrap(err, "GetCodeByCode failed")
		}

		if code == nil {
			code, err = addNewCode(ctx, db, codeR)
		}

		uuidWithHyphen := uuid.New()
		uid = strings.Replace(uuidWithHyphen.String(), "-", "", -1)

		insert := `INSERT INTO ref_usercode ("keyid", "codeid", "userid") VALUES ($1, $2, $3)`

		if _, err := db.ExecContext(ctx, insert, uid, code.CodeID, refUserCode[0].UserID); err != nil {
			return err
		}

		return nil
	}

	if err := RunInTransaction(ctx, api.db, work); err != nil {
		return nil, err
	}

	refUserCode, err := api.GetRefUserCodeByKeyID(ctx, uid)
	if err != nil {
		return nil, errors.Wrap(err, "GetRefUserCodeByKeyID failed")
	}

	return refUserCode, err
}

// UpdateRefUserCode ...
func (api *API) UpdateRefUserCode(ctx context.Context, codeR string, userID string) (*apitypes.RefUserCode, error) { // проверить метод, может не корректно работать !!!

	work := func(ctx context.Context, db TxContext) error {
		refUserCode, err := api.GetRefUserCodeByUserID(ctx, userID)
		if err != nil {
			return errors.Wrapf(err, "SELECT ref_usercode failed: %s", userID)
		}

		code, err := api.GetCodeByCode(ctx, codeR)
		if err != nil {
			return errors.Wrapf(err, "Get code by code failed: %s", codeR)
		}

		if code == nil {
			code, err = api.AddNewCode(ctx, codeR)
			if err != nil {
				return errors.Wrapf(err, "Add new code failed: %s", codeR)
			}
		}

		if refUserCode == nil { // можно убрать, не должно прилетать сюда без установленного ключа
			_, err := api.AddRefUserCode(ctx, codeR, userID)

			if err != nil {
				return errors.Wrapf(err, "Add ref user code failed: %s", userID)
			}
			return nil
		}

		query := `UPDATE ref_usercode SET codeid = $1 WHERE keyid = $2`
		if _, err := db.ExecContext(ctx, query, code.CodeID, refUserCode.KeyID); err != nil {
			return errors.Wrap(err, "UPDATE ref_usercode failed: %s")
		}

		// keyID = refUserCode[0].KeyID
		return nil
	}

	if err := RunInTransaction(ctx, api.db, work); err != nil {
		return nil, err
	}

	refCode, err := api.GetRefUserCodeByUserID(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "Get ref user code by key ID")
	}

	return refCode, err
}

// GetRefUserCodeByKeyID ...
func (api *API) GetRefUserCodeByUserID(ctx context.Context, userID string) (*apitypes.RefUserCode, error) {
	return getRefUserCodeByUserID(ctx, api.db, userID)
}

func getRefUserCodeByUserID(ctx context.Context, db TxContext, userID string) (*apitypes.RefUserCode, error) {
	rowUserCode := []apitypes.RefUserCode{}

	queryRefCode := `SELECT ref_usercode.keyid, ref_usercode.codeid, ref_usercode.userid 
	FROM ref_usercode WHERE ref_usercode.userid = $1 LIMIT 1;`

	err := db.SelectContext(ctx, &rowUserCode, queryRefCode, userID)
	if err != nil {
		return nil, errors.Wrapf(err, "can't get ref_usercode by userid %s", userID)
	}

	if len(rowUserCode) == 0 {
		return nil, nil
	}
	return &rowUserCode[0], nil
}
