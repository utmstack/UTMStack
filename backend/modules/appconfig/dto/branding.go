package dto

import "time"

type BrandingResponse struct {
	Enabled            bool       `json:"enabled"`
	ProductName        string     `json:"productName"`
	LogoURL            string     `json:"logoUrl"`
	LogoDarkURL        string     `json:"logoDarkUrl"`
	FaviconURL         string     `json:"faviconUrl"`
	PrimaryColor       string     `json:"primaryColor"`
	AccentColor        string     `json:"accentColor"`
	LoginBackgroundURL string     `json:"loginBackgroundUrl"`
	FooterText         string     `json:"footerText"`
	SupportURL         string     `json:"supportUrl"`
	HidePoweredBy      bool       `json:"hidePoweredBy"`
	UpdatedAt          *time.Time `json:"updatedAt,omitempty"`
	UpdatedBy          string     `json:"updatedBy,omitempty"`
}

type BrandingRequest struct {
	Enabled            bool   `json:"enabled"`
	ProductName        string `json:"productName"`
	LogoURL            string `json:"logoUrl"`
	LogoDarkURL        string `json:"logoDarkUrl"`
	FaviconURL         string `json:"faviconUrl"`
	PrimaryColor       string `json:"primaryColor"`
	AccentColor        string `json:"accentColor"`
	LoginBackgroundURL string `json:"loginBackgroundUrl"`
	FooterText         string `json:"footerText"`
	SupportURL         string `json:"supportUrl"`
	HidePoweredBy      bool   `json:"hidePoweredBy"`
}

type BrandingPublic struct {
	Enabled            bool   `json:"enabled"`
	ProductName        string `json:"productName"`
	LogoURL            string `json:"logoUrl"`
	LogoDarkURL        string `json:"logoDarkUrl"`
	FaviconURL         string `json:"faviconUrl"`
	PrimaryColor       string `json:"primaryColor"`
	AccentColor        string `json:"accentColor"`
	LoginBackgroundURL string `json:"loginBackgroundUrl"`
	FooterText         string `json:"footerText"`
	SupportURL         string `json:"supportUrl"`
	HidePoweredBy      bool   `json:"hidePoweredBy"`
}
