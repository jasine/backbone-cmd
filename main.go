package main

import (
	"encoding/json"
	"github.com/codegangsta/cli"
	"github.com/deepglint/backbone-cmd/controller"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	//"strconv"
	"bytes"
	"fmt"
	"time"
)

var (
	Version = "0.2.1"
)

type MailConfig struct {
	From         string   `json:"from"`
	To           []string `json:"to"`
	Cc           []string `json:"cc"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	Build_Dir    string   `json:"build_dir"`
	Sub          string   `json:"sub"`
	Content_Path string   `json:"content_path`
	Content      string   `json:"content"`
	Attachment   string   `json:"attach"`
}

///
///
///
///
///
func main() {
	app := cli.NewApp()
	app.Name = "backbone-cli"
	app.Usage = "The Backbone Framework Helper"
	app.Author = "YanHuang"
	app.Email = "yanhuang@deepglint.com"
	app.Commands = []cli.Command{
		{
			Name:  "component",
			Usage: "Components actions",
			Subcommands: []cli.Command{
				{
					Name:   "create",
					Usage:  "Create a Vue Component",
					Action: componentCreate,
				},
			},
		}, {
			Name:  "vulcand",
			Usage: "Generate Vulcand Script",
			Subcommands: []cli.Command{
				{
					Name:   "create",
					Usage:  "yourname componentname vulcandaddr localhostaddr, like :backbone vulcand create huangyan helloworld http://192.168.5.46:8182 http://192.168.1.24:8004",
					Action: vulcandCreate,
				},
			},
		}, {
			Name:  "git",
			Usage: "For github things",
			Subcommands: []cli.Command{
				{
					Name:   "sync",
					Usage:  "make a client to update",
					Action: gitSync,
				}, {
					Name:   "watch",
					Usage:  "always update the github repo",
					Action: gitWatch,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "port,p",
							Value: "8030",
							Usage: "The update port to listen",
						}, cli.StringFlag{
							Name:  "span,s",
							Value: "3",
							Usage: "Sleep time for every update",
						},
					},
				},
			},
		},
	}
	app.Run(os.Args)
}

func gitSync(ctx *cli.Context) {
	if len(ctx.Args()) != 1 {
		log.Println("Error for arg number,please tell the address for remote backbone")
	}
	r, err := http.Get(ctx.Args()[0] + "/pull")
	if err != nil {
		println(err.Error())
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		println(err.Error())
		return
	}
	println(string(b))
}

///
///The Git Watch Helper
///Using /pull to pull the code from repo
///Using /release?tag=... to add the tag and release the code
///
func gitWatch(ctx *cli.Context) {
	go func() {
		for {
			o := execCommand("git", []string{"pull"})
			//i, _ := strconv.ParseInt(ctx.String("span"))
			log.Println(string(o))
			time.Sleep(time.Second * time.Duration(10))
			continue
		}
	}()
	fs := http.FileServer(http.Dir("../build/"))
	//http.Handle("/", fs)
	mux := http.NewServeMux()
	mux.Handle("/", fs)
	mux.HandleFunc("/pull", handleUpdate)
	mux.HandleFunc("/build", handleTag)
	mux.HandleFunc("/release", handleRelease)
	mux.HandleFunc("/liverelease", handleLiveRelease)
	mux.HandleFunc("/install", handleInstall)
	mux.HandleFunc("/ls", handleLs)
	http.ListenAndServe("0.0.0.0:"+ctx.String("port"), mux)
}

func handleLs(w http.ResponseWriter, r *http.Request) {
	o := execCommand("ls", []string{"../build/"})
	w.Write(o)
}

func handleInstall(w http.ResponseWriter, r *http.Request) {
	var ip = r.URL.Query().Get("ip")
	var version = r.URL.Query().Get("tag")

	println(ip)
	println(version)

	if ip == "" || version == "" {
		w.Write([]byte("Please note the ip arg and version arg"))
		return
	}
	pwd, _ := os.Getwd()
	println(pwd)
	//o := execCommand("ssh", []string{"ubuntu@192.168.4.30 'bash -s' < install.sh v0.1.1.test"})

	//o := execCommand("ssh", []string{"ubuntu@192.168.4.30", "'bash -s' < ", pwd + "/install.sh", "v0.1.1.test"})
	o := execCommand("bash", []string{"cmd.sh", ip, version})

	//o := execCommand("ls", []string{"./"})
	w.Write(o)
}

func handleLiveRelease(w http.ResponseWriter, r *http.Request) {
	m, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("Error for Reading Mail Config Json :" + err.Error()))
		return
	}
	var mc MailConfig
	err = json.Unmarshal(m, &mc)
	if err != nil {
		w.Write([]byte("Error for Unmarshalling Mail Config Json :" + err.Error()))
		return
	}

	var t = r.URL.Query().Get("tag")
	println("Getting Tag :" + t)
	mc.Sub = t + " Release"
	c, _ := ioutil.ReadFile(mc.Content_Path)
	mc.Content = string(c)
	mc.Attachment = mc.Build_Dir + t + ".tar.gz"
	if _, err := os.Stat(mc.Attachment); err != nil {
		w.Write([]byte("Error for Checking The AttachMent :" + err.Error()))
		return
	}
	log.Println(mc)
	sendMail2(mc)
	w.Write([]byte("done"))
}

func handleRelease(w http.ResponseWriter, r *http.Request) {
	m, err := ioutil.ReadFile("./mailconfig.json")
	if err != nil {
		w.Write([]byte("Error for Reading Mail Config Json :" + err.Error()))
		return
	}
	var mc MailConfig
	err = json.Unmarshal(m, &mc)
	if err != nil {
		w.Write([]byte("Error for Unmarshalling Mail Config Json :" + err.Error()))
		return
	}

	var t = r.URL.Query().Get("tag")
	println("Getting Tag :" + t)
	mc.Sub = t + " Release"
	c, _ := ioutil.ReadFile(mc.Content_Path)
	mc.Content = string(c)
	mc.Attachment = mc.Build_Dir + t + ".tar.gz"
	if _, err := os.Stat(mc.Attachment); err != nil {
		w.Write([]byte("Error for Checking The AttachMent :" + err.Error()))
		return
	}
	log.Println(mc)
	sendMail2(mc)
	w.Write([]byte("done"))
}
func sendMail2(mc MailConfig) {

	m := gomail.NewMessage()
	m.SetHeader("From", mc.From)
	//m.SetHeader("To", "yanhuang@deepglint.com")
	m.SetHeader("To", mc.To...)
	m.SetHeader("Cc", mc.To...)
	m.SetHeader("Subject", mc.Sub)
	m.SetBody("text/html", mc.Content)

	m.Attach(mc.Attachment)
	log.Println("Here Is in the Send Mail Routine")
	log.Println(mc)
	d := gomail.NewPlainDialer("smtp.exmail.qq.com", 465, mc.Username, mc.Password)

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}
func handleTag(w http.ResponseWriter, r *http.Request) {
	var t = r.URL.Query().Get("tag")
	println("Getting Tag :" + t)
	// o := execCommand("git", []string{"tag", t})
	// o = execCommand("git", []string{"push", "--tag"})
	o := execCommand("bash", []string{"build.sh", t})
	w.Write(o)
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	o := execCommand("git", []string{"pull"})
	w.Write(o)
}

func execCommand(name string, args []string) []byte {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {

		return []byte(fmt.Sprint(err) + ": " + stderr.String())
	}
	return []byte(out.String())
}

func componentCreate(ctx *cli.Context) {
	arg := ctx.Args()
	if len(arg) != 1 {
		log.Println("Please just tell the name of component")
		return
	}
	controller.CreateComponent(arg[0], "./")
}
func vulcandCreate(ctx *cli.Context) {
	arg := ctx.Args()
	if len(arg) != 4 {
		log.Println("Please tell the name of you and the name of component")
		return
	}
	controller.CreateVulcand(arg[0], arg[1], arg[2], arg[3], "./")

}

func sendMail(sub string, content string, attach string, user string, pass string) {

	m := gomail.NewMessage()
	m.SetHeader("From", "backbone@deepglint.com")
	//m.SetHeader("To", "yanhuang@deepglint.com")
	m.SetHeader("To", "weiranyuan@deepglint.com", "zhenyuchen@deepglint.com", "yunhou@deepglint.com", "huiyanliu@deepglint.com", "yanzhang@deepglint.com")
	m.SetHeader("Cc", "libra@deepglint")

	m.SetHeader("Subject", sub)
	m.SetBody("text/html", content)
	m.Attach(attach)
	log.Println(sub, content, attach, user, pass)
	d := gomail.NewPlainDialer("smtp.exmail.qq.com", 465, user, pass)

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}
