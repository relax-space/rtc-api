package cmd

import (
	"fmt"
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
func (d Nginx) Write(p *Project, prefix string) (err error) {

	if len(p.Setting.Ports) == 0 {
		err = fmt.Errorf("port is required,project:%v", p.Name)
		return
	}
	var location string
	location += d.Location(p.Name, p.Setting.Ports[0], prefix)
	for _, sp := range p.Children {
		if len(p.Setting.Ports) == 0 {
			err = fmt.Errorf("port is required,project:%v", sp.Name)
			return
		}
		location += d.Location(sp.Name, sp.Setting.Ports[0], prefix)
	}
	if p.Owner.IsStream {
		location += d.Location(p.Owner.EventProducer.Name, p.Owner.EventProducer.Setting.Ports[0], prefix)
	}
	return (File{}).WriteString(TEMP_FILE+"/nginx", "default.conf", ngnixTemplateServer+location+"\n}")
}

func (Nginx) Location(serverName, port, prefix string) (location string) {
	location = strings.Replace(ngnixTemplateLocation, "$serverName", serverName, -1)
	location = strings.Replace(location, "$containerName", Compose{}.getContainerName(serverName, prefix), -1)
	location = strings.Replace(location, "$port", port, -1)
	return
}
