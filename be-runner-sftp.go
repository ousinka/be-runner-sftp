package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"smart-sftp/conf"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var config *conf.Conf

/**sftp链接
 */
func connect(server conf.RemoteServer) (*sftp.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		sftpClient   *sftp.Client
		err          error
	)
	// get auth method
	if server.IsPrivateKey {
		privateKeyBytes, err := ioutil.ReadFile(server.PrivateKeyFile)
		if err != nil {
			log.Fatal(err)
		}
		key, err := ssh.ParsePrivateKey(privateKeyBytes)
		if err != nil {
			log.Fatal(err)
		}
		auth = []ssh.AuthMethod{ssh.PublicKeys(key)}
	} else {
		auth = make([]ssh.AuthMethod, 0)
		auth = append(auth, ssh.Password(server.Passwd))
	}

	clientConfig = &ssh.ClientConfig{
		User:    server.Username,
		Auth:    auth,
		Timeout: 30 * time.Second,
		//是否免密登录，需要验证服务端，不做验证返回nil
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// connect to ssh
	addr = fmt.Sprintf("%s:%d", server.Host, server.Port)

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create sftp client
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}

	return sftpClient, nil
}

func uploadDirectory(sftpClient *sftp.Client, localPath string, remotePath string, exclude []string) {
	start := time.Now()
	localFiles, err := ioutil.ReadDir(localPath)
	if err != nil {
		log.Println("read dir list fail ", err)
		return
	}

	for _, backupDir := range localFiles {
		localFilePath := path.Join(localPath, backupDir.Name())
		remoteFilePath := path.Join(remotePath, backupDir.Name())
		for _, e := range exclude {
			if strings.HasPrefix(localFilePath, e) {
				log.Println("dir exclude ", e)
				return
			}
		}

		if backupDir.IsDir() {
			sftpClient.Mkdir(remoteFilePath)
			uploadDirectory(sftpClient, localFilePath, remoteFilePath, exclude)
		} else {
			uploadFile(sftpClient, path.Join(localPath, backupDir.Name()), remotePath)
		}
	}

	log.Println(localPath+" copy directory to remote server ", remotePath, " succ. cost:", time.Since(start))
}

func uploadFile(sftpClient *sftp.Client, localFilePath string, remotePath string) {
	start := time.Now()
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		log.Println("os.Open error : ", localFilePath, err)
		return
	}
	defer srcFile.Close()

	var remoteFileName = path.Base(localFilePath)

	var remoteFileFullPath = path.Join(remotePath, remoteFileName)

	if config.SkipExist {
		_, err = sftpClient.Lstat(remoteFileFullPath)
		//为空，说明服务器端已经存在文件，需要跳过
		if err == nil {
			log.Println(remoteFileFullPath, " exist skip.")
			return
		}
	}

	dstFile, err := sftpClient.Create(remoteFileFullPath)
	if err != nil {
		log.Println("sftpClient.Create error : ", path.Join(remotePath, remoteFileName), err)
		return
	}
	defer dstFile.Close()

	ff, err := ioutil.ReadAll(srcFile)
	if err != nil {
		log.Println("ReadAll error : ", localFilePath, err)
		return
	}
	dstFile.Write(ff)
	log.Println(localFilePath+" copy file to remote server ", remotePath, " succ. cost:", time.Since(start))
}

func doUpload(localPath string, argsPath string, server conf.RemoteServer, exclude []string) {
	var (
		err        error
		sftpClient *sftp.Client
	)
	start := time.Now()
	sftpClient, err = connect(server)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer sftpClient.Close()

	_, errStat := sftpClient.Stat(server.Path)
	if errStat != nil {
		sftpClient.Mkdir(server.Path)
		log.Println(server.Path + " remote path not exists. create succ")
	}

	log.Print("start upload dir:", localPath, " to server:", server.Host, " dir:", server.Path)
	if argsPath == "" {
		uploadDirectory(sftpClient, localPath, server.Path, exclude)
	} else {
		uploadDirectory(sftpClient, localPath+argsPath, server.Path+argsPath, exclude)
	}

	elapsed := time.Since(start)
	log.Println("do upload cost: ", elapsed)
}

func main() {
	//记录开始时间
	start := time.Now()
	log.Println("sftp start...")
	config = conf.LoadConf("./conf.json")
	log.Println("config info:", config)

	var argsPath = ""
	if len(os.Args) > 1 {
		argsPath = os.Args[1]
	}

	log.Println("path:", argsPath)
	for _, server := range config.Remote {
		doUpload(config.LocalPath, argsPath, server, config.Exclude)
	}

	log.Println("sync end. cost:", time.Since(start))
}
