"use client";
import Loading from "@/components/Loading";
import { useAuth } from "@/contexts/AuthContext";
import { useSearchParams } from "next/navigation";
import { useEffect } from "react";
import { useTranslations } from "@/contexts/LocaleContext";

export default function LoginPage() {
    const { user, login, loading } = useAuth();
    const { t } = useTranslations();

    const searchParams = useSearchParams();

    const redirectTo = searchParams.get("redirect");

    useEffect(() => {
        if (!loading && !user) {
            login(redirectTo || undefined);
        }
    }, [user, loading, login, redirectTo]);

    if (loading) {
        return <Loading />;
    }

    if (user) {
        return <div>{t("login.already", { name: user.username })}</div>;
    }

    return <div>{t("login.redirecting")}</div>;
}
