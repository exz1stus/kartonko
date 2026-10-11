"use client";
import { useRouter } from "next/navigation";
import { useTranslations } from "@/contexts/LocaleContext";

const LoginButton: React.FC = () => {
    const router = useRouter();
    const { t } = useTranslations();
    return (
        <button
            onClick={() =>
                router.push(
                    `/login?redirect=${encodeURIComponent(window.location.pathname || "/")}`,
                )
            }
            className="w-auto h-10 cursor-pointer"
        >
            <span>{t("nav.login")}</span>
        </button>
    );
};

export default LoginButton;
