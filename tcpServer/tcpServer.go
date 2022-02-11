package main

import (
	"UserController/config"
	"UserController/mysql"
	"UserController/protocol"
	"UserController/redis"
	"UserController/rpc"
	"UserController/utils"
	"log"
)

var Config *config.Config

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	err := config.ReadConfig("..\\config\\config.yaml")
	if err != nil {
		panic(err)
	}

	//connect mysql
	if err = mysql.Connect(); err != nil {
		log.Println(err)
	} else {
		log.Println("mysql connection successed.")
	}

	//connect redis
	if err = redis.Connect(); err != nil {
		log.Println(err)
	} else {
		log.Println("redis connection successed.")
	}
}

func main() {
	Config = config.GetConfig()
	server := rpc.Server()
	server.Register("GetProfile", GetProfile, GetProfileService)
	server.Register("Login", Login, LoginService)
	server.Register("SignUp", SignUp, SignUpService)
	server.Register("UpdateNickName", UpdateNickName, UpdateNickNameService)
	server.Register("UploadProfilePicture", UploadProfilePicture, UploadProfilePictureService)
	server.ListenAndServe(Config.RPC.Address)
}

func GetProfile(v interface{}) interface{} {
	return GetProfileService(*v.(*protocol.ReqGetProfile))
}

func Login(v interface{}) interface{} {
	return LoginService(*v.(*protocol.ReqLogin))
}
func SignUp(v interface{}) interface{} {
	return SignUpService(*v.(*protocol.ReqSignUp))
}
func UpdateNickName(v interface{}) interface{} {
	return UpdateNickNameService(*v.(*protocol.ReqUpdateNickName))
}
func UploadProfilePicture(v interface{}) interface{} {
	return UploadProfilePictureService(*v.(*protocol.ReqUpdateProfilePic))
}
func SignUpService(req protocol.ReqSignUp) (resp protocol.RespSignUp) {
	if req.UserName == "" || req.Password == "" {
		resp.Ret = 1
		return
	}
	if req.NickName == "" {
		req.NickName = req.UserName
	}

	if err := mysql.CreateAccount(req.UserName, req.Password); err != nil {
		resp.Ret = 2
		log.Println("tcp.signUp: mysql.CreateAccount failed. usernam:" + req.UserName + ", err:" + err.Error())
		return
	}
	if err := mysql.CreateProfile(req.UserName, req.NickName); err != nil {
		resp.Ret = 2
		log.Println("tcp.signUp: mysql.CreateProfile failed. usernam:" + req.UserName + ", err:" + err.Error())
		return
	}

	resp.Ret = 0
	return
}
func GetProfileService(req protocol.ReqGetProfile) (resp protocol.RespGetProfile) {
	ok, err := checkToken(req.UserName, req.Token)
	if err != nil {
		resp.Ret = 3
		log.Println("tcp.getProfile: checkToken error." + err.Error())
		return
	}
	if !ok {
		resp.Ret = 1
		log.Println("tcp.getProfile: checkToken failed.")
		return
	}
	nickName, picName, hasData, err := redis.GetProfile(req.UserName)
	if err != nil {
		resp.Ret = 3
		log.Println("tcp.getProfile: checkToken error." + err.Error())
		return
	}
	if hasData {
		log.Println("redis tcp.getProfile done.")
		return protocol.RespGetProfile{Ret: 0, UserName: req.UserName, NickName: nickName, PicName: picName}
	}

	//redis未命中，从MySQL中获取
	nickName, picName, hasData, err = mysql.GetProfile(req.UserName)
	if err != nil {
		resp.Ret = 3
		log.Println("tcp.getProfile: checkToken error." + err.Error())
		return
	}
	if hasData {
		log.Println("mysql tcp.getProfile done.")
		redis.SetNickNameAndPicName(req.UserName, nickName, picName)
	} else {
		resp.Ret = 2
		log.Println("tcp.getProfile: mysql.GetProfile can't find username. username:" + req.UserName)
		return
	}
	log.Println("tcp.getProfile done. username:" + req.UserName)
	return protocol.RespGetProfile{Ret: 0, UserName: req.UserName, NickName: nickName, PicName: picName}
}

func LoginService(req protocol.ReqLogin) (resp protocol.RespLogin) {
	ok, err := mysql.LoginAuth(req.UserName, req.Password)
	if err != nil {
		resp.Ret = 2
		return
	}
	if !ok {
		resp.Ret = 1
		return
	}
	token := utils.GetToken(req.UserName)
	if err := redis.SetToken(req.UserName, token, int64(Config.REDIS.TokenMaxExTime)); err != nil {
		resp.Ret = 2
		return
	}
	resp.Ret = 0
	resp.Token = token
	log.Println("tcp.login: login done.")
	return
}

func UpdateNickNameService(req protocol.ReqUpdateNickName) (resp protocol.RespUpdateNickName) {
	ok, err := checkToken(req.UserName, req.Token)
	if err != nil {
		resp.Ret = 3
		log.Println("tcp.UpdateNickNameService: checkToken error." + err.Error())
		return
	}
	if !ok {
		resp.Ret = 1
		log.Println("tcp.UpdateNickNameService: checkToken failed.")
		return
	}
	if err := redis.InvaildCache(req.UserName); err != nil {
		resp.Ret = 3
		log.Println("tcp.UpdateNickNameService: InvaildCache error." + err.Error())
		return
	}
	ok, err = mysql.UpdateNikcName(req.UserName, req.NickName)
	if err != nil {
		resp.Ret = 3
		log.Println("tcp.UpdateNickNameService: UpdateNikcName error." + err.Error())
		return
	}
	if !ok {
		resp.Ret = 2
		log.Println("tcp.UpdateNickNameService: UpdateNikcName failed.")
		return
	}
	resp.Ret = 0
	log.Println("tcp.UpdateNickNameService done. username:" + req.UserName + ", nickname:" + req.NickName)
	return
}

func UploadProfilePictureService(req protocol.ReqUpdateProfilePic) (resp protocol.RespUpdateProfilePic) {
	ok, err := checkToken(req.UserName, req.Token)
	if err != nil {
		resp.Ret = 3
		log.Println("tcp.UploadProfilePictureService: checkToken error." + err.Error())
		return
	}
	if !ok {
		resp.Ret = 1
		log.Println("tcp.UploadProfilePictureService: checkToken failed.")
		return
	}
	if err := redis.InvaildCache(req.UserName); err != nil {
		resp.Ret = 3
		log.Println("tcp.UploadProfilePictureService: InvaildCache error." + err.Error())
		return
	}
	ok, err = mysql.UploadProfilePicture(req.UserName, req.FileName)
	if err != nil {
		resp.Ret = 3
		log.Println("tcp.UploadProfilePictureService: UploadProfilePictureService error." + err.Error())
		return
	}
	if !ok {
		resp.Ret = 2
		log.Println("tcp.UploadProfilePictureService: UploadProfilePictureService failed.")
		return
	}
	resp.Ret = 0
	log.Println("tcp.UploadProfilePictureService done. username:" + req.UserName + ", picname:" + req.FileName)
	return
}

func checkToken(userName string, token string) (bool, error) {
	return redis.CheckToken(userName, token)
}
