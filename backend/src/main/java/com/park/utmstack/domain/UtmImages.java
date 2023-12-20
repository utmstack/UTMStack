package com.park.utmstack.domain;


import com.park.utmstack.domain.shared_types.enums.ImageShortName;

import javax.persistence.*;
import java.io.Serializable;

/**
 * A UtmImages.
 */
@Entity
@Table(name = "utm_images")
public class UtmImages implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @Enumerated(EnumType.STRING)
    @Column(name = "short_name")
    private ImageShortName shortName;

    @Column(name = "tooltip_text")
    private String tooltipText;

    @Column(name = "system_img")
    private String systemImg;

    @Column(name = "user_img")
    private String userImg;

    public ImageShortName getShortName() {
        return shortName;
    }

    public UtmImages shortName(ImageShortName shortName) {
        this.shortName = shortName;
        return this;
    }

    public void setShortName(ImageShortName shortName) {
        this.shortName = shortName;
    }

    public String getTooltipText() {
        return tooltipText;
    }

    public UtmImages tooltipText(String tooltipText) {
        this.tooltipText = tooltipText;
        return this;
    }

    public void setTooltipText(String tooltipText) {
        this.tooltipText = tooltipText;
    }


    public String getSystemImg() {
        return systemImg;
    }

    public void setSystemImg(String systemImg) {
        this.systemImg = systemImg;
    }

    public UtmImages systemImg(String systemImg) {
        this.systemImg = systemImg;
        return this;
    }

    public String getUserImg() {
        return userImg;
    }

    public UtmImages userImg(String userImg) {
        this.userImg = userImg;
        return this;
    }

    public void setUserImg(String userImg) {
        this.userImg = userImg;
    }
}
