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

// GetUserByName ...
func (api *API) GetUserByName(nameUser string) (*apitypes.UserRow, error) {
	return getUserByName(context.Background(), api.db, nameUser)
}

func getUserByName(ctx context.Context, db TxContext, nameUser string) (*apitypes.UserRow, error) {
	userRow := []apitypes.UserRow{}

	err := db.SelectContext(ctx, &userRow, "SELECT * FROM prj_user WHERE (nameuser = $1) LIMIT 1;", nameUser)

	if err != nil {
		return nil, errors.Wrapf(err, "can't get user by name %s", nameUser)
	}
	if len(userRow) == 0 {
		return nil, err
	}
	return &userRow[0], nil
}

// GetUserByID ...
func (api *API) GetUserByID(idUser string) (*apitypes.UserRow, error) {
	return getUserByID(context.Background(), api.db, idUser)
}

func getUserByID(ctx context.Context, db TxContext, userID string) (*apitypes.UserRow, error) {
	userRow := []apitypes.UserRow{}

	err := db.SelectContext(ctx, &userRow, "SELECT * FROM prj_user WHERE (userid = $1) LIMIT 1;", userID)
	if err != nil {
		return nil, errors.Wrap(err, "getUserById failed")
	}

	if len(userRow) == 0 {
		return nil, err
	}

	return &userRow[0], err
}

// GetCodeByID ...
func (api *API) GetCodeByID(idCode string) (*apitypes.CodeRow, error) {
	return getCodeByID(context.Background(), api.db, idCode)
}

func getCodeByID(ctx context.Context, db TxContext, idCode string) (*apitypes.CodeRow, error) {
	const query = `SELECT * 
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
func (api *API) GetCodeByCode(codeCode string) (*apitypes.CodeRow, error) {
	return getCodeByCode(context.Background(), api.db, codeCode)
}

func getCodeByCode(ctx context.Context, db TxContext, codeCode string) (*apitypes.CodeRow, error) {
	codeRow := []apitypes.CodeRow{}

	err := db.SelectContext(ctx, &codeRow, "SELECT * FROM prj_code WHERE (code = $1)", codeCode)
	if err != nil {
		return nil, errors.Wrap(err, "GetCodeByCode api.db.Select failed with an error")
	}

	if len(codeRow) == 1 {
		return &codeRow[0], err
	}

	return nil, err
}

// GetRefUserCodeByKeyID ...
func (api *API) GetRefUserCodeByKeyID(keyID string) (*apitypes.RefUserCode, error) {
	return getRefUserCodeByKeyID(context.Background(), api.db, keyID)
}

func getRefUserCodeByKeyID(ctx context.Context, db TxContext, keyID string) (*apitypes.RefUserCode, error) {
	rowUserCode := []apitypes.RefUserCode{}
	err := db.SelectContext(ctx, &rowUserCode, "SELECT * FROM ref_usercode WHERE keyid = $1 LIMIT 1;", keyID)
	if err != nil {
		return nil, errors.Wrapf(err, "can't get ref_usercode by keyid %s", keyID)
	}

	if len(rowUserCode) == 0 {
		return nil, nil
	}
	return &rowUserCode[0], nil
}

// GetRefUserCodeByUserName ...
func (api *API) GetRefUserCodeByUserName(userN string) (*apitypes.RefUserCode, error) {
	return getRefUserCodeByUserName(context.Background(), api.db, userN)
}

func getRefUserCodeByUserName(ctx context.Context, db TxContext, userN string) (*apitypes.RefUserCode, error) {
	user, err := getUserByName(context.Background(), db, userN) // использовать context.Background() или ctx ?

	if err != nil {
		return nil, errors.Wrapf(err, "can't get user by name %s", userN)
	}

	if user == nil {
		return nil, errors.Wrap(err, "Такой пользователя не зарегестрирован")
	}

	rowUserCode := []apitypes.RefUserCode{}
	err = db.SelectContext(ctx, &rowUserCode, "SELECT * FROM ref_usercode WHERE userid = $1 LIMIT 1;", string(user.UserID))
	if err != nil {
		return nil, errors.Wrapf(err, "can't get ref_usercode by uer id %s", string(user.UserID))
	}

	if len(rowUserCode) == 0 {
		return nil, nil
	}

	return &rowUserCode[0], nil
}

// AddNewUser ...
func (api *API) AddNewUser(userN, userID, chatID string) (*apitypes.UserRow, error) {
	user, err := api.GetUserByID(userID)

	if err != nil {
		return nil, errors.Wrap(err, "AddNewUser GetUserByID failed with an error")
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
		return nil, errors.Wrap(err, "AddNewUser GetUserByID failed with an error")
	}

	return user, err
}

// AddNewCode ...
func (api *API) AddNewCode(codeN string) (*apitypes.CodeRow, error) {
	var code *apitypes.CodeRow
	var err error

	work := func(ctx context.Context, db TxContext) error {
		code, err = addNewCode(ctx, db, codeN)
		if err != nil {
			return err
		}
		return nil
	}

	if err := RunInTransaction(context.Background(), api.db, work); err != nil {
		return nil, err
	}

	return code, nil
}

func addNewCode(ctx context.Context, db TxContext, codeN string) (*apitypes.CodeRow, error) {
	var code *apitypes.CodeRow
	var err error
	var uid string

	code, err = getCodeByCode(context.Background(), db, codeN)
	if err != nil {
		return nil, errors.Wrap(err, "AddNewCode GetCodeByCode failed with an error")
	}

	if code != nil {
		return nil, err
	}

	uuidWithHyphen := uuid.New()
	uid = strings.Replace(uuidWithHyphen.String(), "-", "", -1)

	if _, err := db.ExecContext(ctx, `INSERT INTO prj_code ("codeid", "code") VALUES ($1, $2)`, uid, codeN); err != nil {
		return nil, err
	}

	code, err = getCodeByID(context.Background(), db, uid)
	if err != nil {
		return nil, errors.Wrap(err, "AddNewCode GetCodeByID failed with an error")
	}

	return code, nil
}

// AddRefUserCode ...
func (api *API) AddRefUserCode(codeR string, userIDR string) (*apitypes.RefUserCode, error) {
	var refUserCode *apitypes.RefUserCode
	var err error
	var uid string

	user, err := getUserByID(context.Background(), api.db, userIDR)
	if err != nil {
		return nil, errors.Wrap(err, "get user by Id failed")
	}
	if user == nil {
		return nil, errors.Wrap(err, "Такой пользователя не зарегестрирован")
	}

	work := func(ctx context.Context, db TxContext) error {
		refUserCode, err := getRefUserCodeByUserName(context.Background(), api.db, user.NameUser) // использовать context.Background() или ctx ?
		if err != nil {
			return errors.Wrap(err, "getRefUserCodeByUserName failed")
		}

		if refUserCode != nil {
			log.Printf("Пользователь уже установил кодовое слово")
			return err
		}

		code, err := getCodeByCode(context.Background(), api.db, codeR) // использовать context.Background() или ctx ?
		if err != nil {
			return errors.Wrap(err, "GetCodeByCode failed")
		}

		if code == nil {
			code, err = addNewCode(ctx, db, codeR)
		}

		uuidWithHyphen := uuid.New()
		uid = strings.Replace(uuidWithHyphen.String(), "-", "", -1)

		if _, err := db.ExecContext(ctx, `INSERT INTO ref_usercode ("keyid", "codeid", "userid") VALUES ($1, $2, $3)`, uid, code.CodeID, user.UserID); err != nil {
			return err
		}

		return nil
	}

	if err := RunInTransaction(context.Background(), api.db, work); err != nil {
		return nil, err
	}

	refUserCode, err = api.GetRefUserCodeByKeyID(uid)
	if err != nil {
		return nil, errors.Wrap(err, "GetRefUserCodeByKeyID failed")
	}

	return refUserCode, err
}

// UpdateRefUserCode ...
func (api *API) UpdateRefUserCode(codeR string, userR string) (*apitypes.RefUserCode, error) { // проверить метод, может не корректно работать !!!

	var refUserCode *apitypes.RefUserCode
	var keyID string

	work := func(ctx context.Context, db TxContext) error {
		user, err := api.GetUserByName(userR)
		if err != nil {
			return errors.Wrapf(err, "Get user by name failed: %s", userR)
		}

		refUserCode, err := api.GetRefUserCodeByUserName(user.NameUser)
		if err != nil {
			return errors.Wrapf(err, "Get ref user code by user name failed: %s", user.NameUser)
		}

		code, err := api.GetCodeByCode(codeR)
		if err != nil {
			return errors.Wrapf(err, "Get code by code failed: %s", codeR)
		}

		if code == nil {
			code, err = api.AddNewCode(codeR)
		}

		if err != nil {
			return errors.Wrapf(err, "Add new code failed: %s", codeR)
		}

		if refUserCode == nil {
			refUserCode, err = api.AddRefUserCode(codeR, userR)

			if err != nil {
				return errors.Wrapf(err, "Add ref user code failed: %s", userR)
			}
		}
		keyID = refUserCode.KeyID

		if _, err := db.ExecContext(ctx, `UPDATE ref_usercode SET codeid = $1 WHERE keyid = $2`, code.CodeID, refUserCode.KeyID); err != nil {
			return errors.Wrap(err, "UPDATE ref_usercode failed: %s")
		}

		return nil
	}

	if err := RunInTransaction(context.Background(), api.db, work); err != nil {
		return nil, err
	}

	refUserCode, err := api.GetRefUserCodeByKeyID(keyID)
	if err != nil {
		return nil, errors.Wrap(err, "Get ref user code by key ID")
	}

	return refUserCode, err
}
