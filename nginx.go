package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	ngnixTemplateServer = `server {
		listen       80;
		server_name  test.local.com;
		location / {
			root   /usr/share/nginx/html;
			index  index.html index.htm;
		}
		`
	ngnixTemplateLocation = `location /$serverName/ {
		proxy_set_header Host $host;
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header X-Forwarded-Proto $scheme;
		proxy_set_header Connection keep-alive;
		proxy_pass       http://$containerName:$port/;
	}
	`
)

type Nginx struct {
}

// setNgnix set nginx default.conf
func (d Nginx) WriteConfig(p *ProjectDto, eventBrokerPort string) (err error) {

	if len(p.Ports) == 0 {
		err = fmt.Errorf("port is required,project:%v", p.ServiceName)
		return
	}
	var location string
	location += d.Location(p.ServiceName, p.Ports[0])

	for _, sp := range p.SubProjects {
		if len(p.Ports) == 0 {
			err = fmt.Errorf("port is required,project:%v", sp.ServiceName)
			return
		}
		location += d.Location(sp.ServiceName, sp.Ports[0])
	}

	if (ProjectInfo{}).ShouldEventBroker(p) {
		location += d.Location(EventBroker_Name, eventBrokerPort)
	}
	if err = os.MkdirAll(TEMP_FILE+"/nginx", os.ModePerm); err != nil {
		return
	}
	return (File{}).WriteString(TEMP_FILE+"/nginx/default.conf", ngnixTemplateServer+location+"\n}")
}

func (Nginx) Location(serverName, port string) (location string) {
	location = strings.Replace(ngnixTemplateLocation, "$serverName", serverName, -1)
	location = strings.Replace(location, "$containerName", Compose{}.getContainerName(serverName), -1)
	location = strings.Replace(location, "$port", port, -1)
	return
}
