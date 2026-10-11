"use client";

import Dropdown from "@/components/template/Dropdown";
import { useTranslations } from "@/contexts/LocaleContext";
import Link from "next/link";

export default function SidebarLinks({ username, authenticated, moderator }: { username?: string; authenticated: boolean; moderator: boolean }) {
    const { t } = useTranslations();
    return <>
        <Link href="/">{t("nav.gallery")}</Link>
        <Link href="/boards">{t("nav.boards")}</Link>
        <Link href="/tags">{t("nav.tags")}</Link>
        {authenticated && <><Link href={`/user/${username}`}>{t("nav.profile")}</Link><Link href="/upload">{t("nav.upload")}</Link></>}
        {moderator && <Dropdown open={true}><Link href="/log">{t("nav.log")}</Link></Dropdown>}
    </>;
}
