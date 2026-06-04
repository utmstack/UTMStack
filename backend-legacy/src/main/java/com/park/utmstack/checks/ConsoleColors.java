package com.park.utmstack.checks;

public class ConsoleColors {
    public static void reset() {
        System.out.print("\033[0m");
    }

    public static void redBold() {
        System.out.print("\033[1;31m");
    }

    public static void cyanBold() {
        System.out.print("\033[1;36m");
    }

    public static void greenBold() {
        System.out.print("\033[1;32m");
    }

    public static void magentaBold() {
        System.out.print("\033[1;35m");
    }

    public static void yellowBold() {
        System.out.print("\033[1;33m");
    }
}
