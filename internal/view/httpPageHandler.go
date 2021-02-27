package view

import (
	"fmt"
	"github.com/githubchry/goweb/internal/dao/models"
	"github.com/githubchry/goweb/internal/logics/protos"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
)

func HTTPUserSettingPageHandler(w http.ResponseWriter, r *http.Request) {

	// 查询用户名是否已存在
	var result protos.User
	models.NewMgo().FindOne("username", mux.Vars(r)["username"]).Decode(&result)

	if len(result.Username) <= 0 {
		fmt.Fprintf(w, "用户不存在!\n")
		return
	}
	// 解析指定文件生成模板对象
	tmpl, err := template.ParseFiles("./www/template/settings.tmpl")
	if err != nil {
		fmt.Println("create template failed, err:", err)
		return
	}

	var tmplUserPage struct {
		Username string
		Email    string
		Photo    string
	}
	tmplUserPage.Username = result.Username
	tmplUserPage.Email = result.Email
	if len(result.Photo) > 0 {
		tmplUserPage.Photo = models.PreDownload("photo", result.Photo)
	}

	// 利用给定数据渲染模板，并将结果写入w
	tmpl.Execute(w, tmplUserPage)
}

//[Go语言标准库之template](https://www.cnblogs.com/nickchen121/p/11517448.html)
//[GO Web编程示例 - 路由（使用gorilla/mux）](https://www.jianshu.com/p/698156c07ad4)
//[golang模板语法简明教程](https://www.cnblogs.com/Pynix/p/4154630.html)
func HTTPUserPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(mux.Vars(r)["username"])
	log.Println("method:", r.Method) //获取请求的方法

	// 解析指定文件生成模板对象
	tmpl, err := template.ParseFiles("./www/template/user.tmpl")
	if err != nil {
		fmt.Println("create template failed, err:", err)
		return
	}

	// 查询用户名是否已存在
	var result protos.User
	models.NewMgo().FindOne("username", mux.Vars(r)["username"]).Decode(&result)

	var tmplUser struct {
		Username string
		Email    string
		Photo    string
	}
	tmplUser.Username = result.Username
	tmplUser.Email = result.Email
	tmplUser.Photo = models.PreDownload("photo", result.Photo)

	// 利用给定数据渲染模板，并将结果写入w
	tmpl.Execute(w, tmplUser)
}
