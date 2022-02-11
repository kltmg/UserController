package main

import (
	"UserController/config"
	"UserController/protocol"
	"UserController/rpc"
	"UserController/utils"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"
)

var loginTemplate *template.Template
var profileTemplate *template.Template
var jumpTemplate *template.Template
var signupTemplate *template.Template

type LoginResponse struct {
	Msg string
}
type ProfileResponse struct {
	UserName string
	NickName string
	PicName  string
}
type JumpResponse struct {
	Msg string
}

type SignupResponse struct {
	Msg string
}

var rpcClient rpc.RPCClient
var Config *config.Config

// init 提前解析html文件.程序用到即可直接使用，避免多次解析
func init() {
	loginTemplate = template.Must(template.ParseFiles("../templates/login.html"))
	profileTemplate = template.Must(template.ParseFiles("../templates/profile.html"))
	jumpTemplate = template.Must(template.ParseFiles("../templates/jump.html"))
	signupTemplate = template.Must(template.ParseFiles("../templates/signup.html"))
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)

	err := config.ReadConfig("..\\config\\config.yaml")
	if err != nil {
		panic(err)
	}
}

func main() {
	Config = config.GetConfig()
	var err error
	rpcClient, err = rpc.Client(Config.RPC.ConnPoolSize, Config.RPC.Address)
	if err != nil {
		panic(err)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(Config.HTTP.StaticFilePath))))
	http.HandleFunc("/", GetProfile)
	http.HandleFunc("/signUp", SignUp)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/profile", GetProfile)
	http.HandleFunc("/updateNickName", UpdateNickName)
	http.HandleFunc("/uploadFile", UploadProfilePicture)

	if err := http.ListenAndServe(":"+Config.HTTP.Port, nil); err != nil {
		panic(err)
	}
}

func GetProfile(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		token, err := req.Cookie("token")
		if err != nil {
			log.Println(err)
			templateLogin(rw, LoginResponse{Msg: ""})
			return
		}
		username := req.FormValue("username")
		if username == "" {
			nameCookie, err := req.Cookie("username")
			if err != nil {
				log.Println(err)
				templateLogin(rw, LoginResponse{Msg: ""})
				return
			}
			username = nameCookie.Value
		}

		req := protocol.ReqGetProfile{
			UserName: username,
			Token:    token.Value,
		}
		resp := protocol.RespGetProfile{}
		if err := rpcClient.Call("GetProfile", req, &resp); err != nil {
			log.Println("http.GetProfile: Call GetProfile failed. username:" + username + ", err:" + err.Error())
			templateJump(rw, JumpResponse{Msg: "获取用户信息失败!"})
			return
		}

		switch resp.Ret {
		case 0:
			if resp.PicName == "" {
				resp.PicName = Config.DefaultImagePath
			}
			templateProfile(rw, ProfileResponse{
				UserName: resp.UserName,
				NickName: resp.NickName,
				PicName:  resp.PicName,
			})
		case 1:
			templateLogin(rw, LoginResponse{Msg: "请重新登陆!"})
		case 2:
			templateJump(rw, JumpResponse{Msg: "用户不存在!"})
		default:
			templateJump(rw, JumpResponse{Msg: "获取用户信息失败!"})
		}
		log.Println("http.GetProfile: GetProfile done. username:" + username + ", ret:" + strconv.Itoa(resp.Ret))
	}
}

func SignUp(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		username := req.FormValue("username")
		nickname := req.FormValue("nickname")
		password := req.FormValue("password")

		if username == "" || password == "" {
			templateSignup(rw, SignupResponse{Msg: "用户名/密码不能为空!"})
			return
		}
		log.Println("http: signup username:" + username + ", nickname:" + nickname + ", password:" + password)
		req := protocol.ReqSignUp{
			UserName: username,
			NickName: nickname,
			Password: password,
		}
		resp := protocol.RespSignUp{}
		if err := rpcClient.Call("SignUp", req, &resp); err != nil {
			log.Println("http.SignUp: Call SignUp failed. username:" + username + ", err:" + err.Error())
			templateSignup(rw, SignupResponse{Msg: "注册失败!"})
			return
		}
		switch resp.Ret {
		case 0:
			templateLogin(rw, LoginResponse{Msg: "创建账号成功，请登陆!"})
		case 1:
			templateSignup(rw, SignupResponse{Msg: "用户名/密码不能为空!"})
		case 2:
			templateSignup(rw, SignupResponse{Msg: "创建账号失败!"})
		}
		log.Println("http.SignUp: SignUp done.")
	}
	if req.Method == "GET" {
		templateSignup(rw, SignupResponse{Msg: ""})
	}
}
func Login(rw http.ResponseWriter, req *http.Request) {
	// 可以用req.URL.Query().Get("xxx")来获取GET参数
	if req.Method == "POST" {
		username := req.FormValue("username")
		password := req.FormValue("password")
		if username == "" || password == "" {
			templateLogin(rw, LoginResponse{Msg: "用户名/密码不能为空!"})
			return
		}
		req := protocol.ReqLogin{
			UserName: username,
			Password: password,
		}
		resp := protocol.RespLogin{}
		if err := rpcClient.Call("Login", req, &resp); err != nil {
			log.Println("http.Login: Call Login failed. username:" + username + ", err:" + err.Error())
			templateLogin(rw, LoginResponse{Msg: "登陆失败!"})
			return
		}
		switch resp.Ret {
		case 0:
			cookie := http.Cookie{Name: "username", Value: username, MaxAge: Config.REDIS.TokenMaxExTime}
			http.SetCookie(rw, &cookie)
			cookie = http.Cookie{Name: "token", Value: resp.Token, MaxAge: Config.REDIS.TokenMaxExTime}
			http.SetCookie(rw, &cookie)
			templateJump(rw, JumpResponse{"登陆成功"})
		case 1:
			templateLogin(rw, LoginResponse{Msg: "用户名或密码错误!"})
		default:
			templateLogin(rw, LoginResponse{Msg: "登陆失败!"})
		}
		log.Println("http.Login: Login done. username:" + username)
	}

}
func UpdateNickName(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		token, err := req.Cookie("token")
		if err != nil {
			log.Println(err)
			templateLogin(rw, LoginResponse{Msg: ""})
			return
		}

		username := req.FormValue("username")
		nickname := req.FormValue("nickname")

		req := protocol.ReqUpdateNickName{
			UserName: username,
			NickName: nickname,
			Token:    token.Value,
		}
		resp := protocol.RespUpdateNickName{}

		if err := rpcClient.Call("UpdateNickName", req, &resp); err != nil {
			log.Println("http.UpdateNickName: Call UpdateNickName failed. username:" + username + ", err:" + err.Error())
			templateJump(rw, JumpResponse{Msg: "修改昵称失败!"})
			return
		}
		switch resp.Ret {
		case 0:
			templateJump(rw, JumpResponse{Msg: "修改昵称成功"})
		case 1:
			templateLogin(rw, LoginResponse{Msg: "请重新登陆"})
		case 2:
			templateJump(rw, JumpResponse{Msg: "用户不存在"})
		default:
			templateJump(rw, JumpResponse{Msg: "修改昵称失败"})
		}
		log.Println("http.UpdateNickName: Call UpdateNickName done. username:" + username + ", nickname:" + nickname + ", ret:" + strconv.Itoa(resp.Ret))
	}
}
func UploadProfilePicture(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		token, err := req.Cookie("token")
		if err != nil {
			log.Println(err)
			templateLogin(rw, LoginResponse{Msg: ""})
			return
		}
		username := req.FormValue("username")
		file, head, err := req.FormFile("image")
		if err != nil {
			log.Println("http.UploadProfilePicture: Call UploadProfilePicture failed. username:" + username + ", err:" + err.Error())
			templateJump(rw, JumpResponse{Msg: "获取图片失败!"})
			return
		}
		defer file.Close()
		newName, isLegal := utils.CheckAndCreateFileName(head.Filename)
		if !isLegal {
			templateJump(rw, JumpResponse{Msg: "文件格式不支持!"})
			return
		}
		filePath := Config.HTTP.StaticFilePath + newName
		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
		defer dstFile.Close()
		//拷贝文件.
		_, err = io.Copy(dstFile, file)
		if err != nil {
			templateJump(rw, JumpResponse{Msg: "文件拷贝出错!"})
			return
		}

		req := protocol.ReqUpdateProfilePic{
			UserName: username,
			FileName: newName,
			Token:    token.Value,
		}
		resp := protocol.RespUpdateProfilePic{}
		if err := rpcClient.Call("UploadProfilePicture", req, &resp); err != nil {
			log.Println("http.UploadProfilePicture: Call UploadProfilePicture failed. username:" + username + ", err:" + err.Error())
			templateJump(rw, JumpResponse{Msg: "修改头像失败!"})
			return
		}
		switch resp.Ret {
		case 0:
			templateJump(rw, JumpResponse{Msg: "修改头像成功"})
		case 1:
			templateLogin(rw, LoginResponse{Msg: "请重新登陆"})
		case 2:
			templateJump(rw, JumpResponse{Msg: "用户不存在"})
		default:
			templateJump(rw, JumpResponse{Msg: "修改头像失败"})
		}
		log.Println("http.UploadProfilePicture: Call UploadProfilePicture done. username:" + username + ", picname:" + newName + ", ret:" + strconv.Itoa(resp.Ret))
	}
}

func templateLogin(rw http.ResponseWriter, resp LoginResponse) {
	if err := loginTemplate.Execute(rw, resp); err != nil {
		log.Println(err)
	}
}

func templateProfile(rw http.ResponseWriter, resp ProfileResponse) {
	if err := profileTemplate.Execute(rw, resp); err != nil {
		log.Println(err)
	}
}

func templateJump(rw http.ResponseWriter, resp JumpResponse) {
	if err := jumpTemplate.Execute(rw, resp); err != nil {
		log.Println(err)
	}
}

func templateSignup(rw http.ResponseWriter, resp SignupResponse) {
	if err := signupTemplate.Execute(rw, resp); err != nil {
		log.Println(err)
	}
}
