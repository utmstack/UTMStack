package network

import (
	"path"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/docker"
	"github.com/utmstack/UTMStack/installer/templates"
	"github.com/utmstack/UTMStack/installer/utils"
)

type NginxConfig struct {
	SharedKey string
}

func InstallNginx(distro string) error {
	switch distro {
	case config.RequiredDistroUbuntu:
		env := []string{"DEBIAN_FRONTEND=noninteractive"}

		if err := utils.RunEnvCmd(env, "apt-get", "install", "-y", "nginx"); err != nil {
			return err
		}
	case config.RequiredDistroRHEL:
		env := []string{"DNF_YUM_AUTO_YES=1"}

		if err := utils.RunEnvCmd(env, "dnf", "install", "-y", "nginx"); err != nil {
			return err
		}
	}
	return nil
}

func ConfigureNginx(conf *config.Config, stack *docker.StackConfig, distro string) error {
	c := NginxConfig{
		SharedKey: conf.InternalKey,
	}

	err := utils.GenerateConfig(c, templates.FrontEnd, path.Join(stack.FrontEndNginx, "00_nginx_panel.conf"))
	if err != nil {
		return err
	}

	err = utils.WriteToFile(path.Join("/", "etc", "nginx", "html", "custom_502.html"), templates.NginxCustomBadGateway)
	if err != nil {
		config.Logger().ErrorF("error writing custom 502 page: %v", err)
	}

	switch distro {
	case config.RequiredDistroUbuntu:
		err = utils.GenerateConfig(c, templates.ProxyUbuntu, path.Join("/", "etc", "nginx", "sites-available", "default"))
		if err != nil {
			return err
		}

	case config.RequiredDistroRHEL:
		err = utils.GenerateConfig(c, templates.ProxyRHEL, path.Join("/", "etc", "nginx", "nginx.conf"))
		if err != nil {
			return err
		}

		if err := utils.RunCmd("chcon", "-R", "-t", "httpd_sys_content_t", "/utmstack/cert/"); err != nil {
			return err
		}

		if err := utils.RunCmd("setsebool", "-P", "httpd_read_user_content", "1"); err != nil {
			return err
		}
	}

	if err := utils.RunCmd("systemctl", "restart", "nginx"); err != nil {
		return err
	}

	return nil
}
