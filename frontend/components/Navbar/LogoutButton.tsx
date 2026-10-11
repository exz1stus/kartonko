import { useAuth } from "@/contexts/AuthContext";
import React from "react";
import { useTranslations } from "@/contexts/LocaleContext";

const LogoutButton: React.FC = () => {
    const { logout } = useAuth();
    const { t } = useTranslations();

    return (
        <button
            className="bg-surface-10 px-2 border hover:border-surface-30 rounded-xl w-full cursor-pointer"
            onClick={async () => await logout()}
        >
            {t("nav.logout")}
        </button>
    );
};

export default LogoutButton;
