package main

import (
	"path"

	"github.com/utmstack/UTMStack/installer/templates"
	"github.com/utmstack/UTMStack/installer/types"
	"github.com/utmstack/UTMStack/installer/utils"
)

func InstallNginx(distro string) error {
	switch distro {
	case "ubuntu":
		env := []string{"DEBIAN_FRONTEND=noninteractive"}

		if err := utils.RunEnvCmd(env, "apt-get", "update"); err != nil {
			return err
		}

		if err := utils.RunEnvCmd(env, "apt-get", "install", "-y", "nginx"); err != nil {
			return err
		}

	case "redhat":
		env := []string{"DNF_YUM_AUTO_YES=1"}

		if err := utils.RunEnvCmd(env, "dnf", "install", "-y", "nginx"); err != nil {
			return err
		}
	}
	return nil
}

func ConfigureNginx(conf *types.Config, stack *types.StackConfig, distro string) error {
	c := types.NginxConfig{
		SharedKey: conf.InternalKey,
	}

	err := utils.GenerateConfig(c, templates.FrontEnd, path.Join(stack.FrontEndNginx, "00_nginx_panel.conf"))
	if err != nil {
		return err
	}

	switch distro {
	case "ubuntu":
		err = utils.GenerateConfig(c, templates.ProxyUbuntu, path.Join("/", "etc", "nginx", "sites-available", "default"))
		if err != nil {
			return err
		}

	case "redhat":
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
