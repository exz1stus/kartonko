"use client";

import { useTranslations } from "@/contexts/LocaleContext";
import TimeField from "@/components/TimeField";

export default function ProfileMetadata({ privilege, joinedAt, lastSeen }: { privilege: string; joinedAt: string; lastSeen: string }) {
    const { t } = useTranslations();
    return <>
        <div className="flex justify-between"><span>{t("profile.privilege")}</span><span>{privilege}</span></div>
        <div className="flex justify-between"><span>{t("profile.joined")}</span><TimeField time={joinedAt} /></div>
        <div className="flex justify-between"><span>{t("profile.lastSeen")}</span><TimeField time={lastSeen} /></div>
    </>;
}
