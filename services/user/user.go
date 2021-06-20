package user

import (
	"errors"
	"fmt"
	. "gen/models"
	"gen/registry"
	. "gen/services/sql_store"
	"gen/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"time"
)

var (
	Secret        = []byte("@fc6951544^f55c644!@0d")
	Existed       = errors.New("邮箱已存在")
	NotExisted    = errors.New("邮箱不存在")
	PasswordWrong = errors.New("邮箱或密码错误")
	LoginFailed   = errors.New("登录失败")
)

type Service struct {
	SQLStore *SQLStore `inject:""`
}

func init() {
	registry.RegisterService(&Service{})
}

func (r Service) Init() error {
	return nil
}

func (r Service) Register(name string, email string, password string) (string, error) {
	emailExisted, err := IsUserEmailExisted(email)
	if err != nil {
		return "", err
	}
	if emailExisted {
		return "", Existed
	}
	var user = User{}
	salt := utils.GetUuidV4()[24:]
	user.Name = name
	user.Email = email
	user.Password = utils.Sha1([]byte(password + salt))
	user.Salt = salt
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	err = r.SQLStore.DB().Save(&user).Error
	if err != nil {
		return "", err
	} else {
		token, err := r.createToken(user.ID)
		if err != nil {
			return "", err
		} else {
			return token, nil
		}
	}
}

func (r Service) Login(email string, password string) (string, error) {
	var user User
	err := r.SQLStore.DB().Where("email = ?", email).First(&user).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", NotExisted
		}
		return "", err
	}
	if user.Password != utils.Sha1([]byte(password+user.Salt)) {
		return "", PasswordWrong
	} else {
		token, err := r.createToken(user.ID)
		if err != nil {
			return "", LoginFailed
		} else {
			return token, nil
		}
	}
}

// ParseToken 解析token
func (r Service) ParseToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return Secret, nil
	})
	if err != nil {
		return 0, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return uint(claims["userId"].(float64)), nil
	} else {
		return 0, err
	}
}

// 创建token
func (r Service) createToken(userId uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString(Secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}