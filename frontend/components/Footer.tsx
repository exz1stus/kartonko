"use client";
import Link from "next/link";
import { useTranslations } from "@/contexts/LocaleContext";
import { Palette } from "lucide-react";
import { OPEN_APPEARANCE_SETTINGS_EVENT } from "@/lib/settingsEvents";

const Footer = () => {
    const { t } = useTranslations();

    return (
        <div className="interface-chrome flex flex-row justify-end items-center gap-4 bg-surface-0/80 px-4 border-t w-full glass">
            <button type="button" aria-label={t("settings.appearance")} onClick={() => window.dispatchEvent(new Event(OPEN_APPEARANCE_SETTINGS_EVENT))} className="inline-flex items-center justify-center rounded-md p-1.5 hover:bg-surface-20 focus-visible:outline-2 focus-visible:outline-primary-0">
                <Palette size={18} />
            </button>
            <span className="inline-flex">
                <Link href="/about">{t("nav.about")}</Link>
            </span>
            <span className="inline-flex gap-1">
                <Link href="https://github.com/exz1stus/kartonko">GitHub</Link>
            </span>
        </div>
    );
};

export default Footer;
