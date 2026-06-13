"use client";
import { useAuth } from "@/contexts/AuthContext";
import { useRouter } from "next/navigation";
import { ReactNode, useEffect } from "react";
import Loading from "./Loading";

interface Props {
    moderator?: boolean;
    children: ReactNode;
}

const AuthGuard = ({ moderator = false, children }: Props) => {
    const { user, loading } = useAuth();
    const router = useRouter();
    useEffect(() => {
        if (loading) return;
        if (!user) {
            router.replace(
                `/login?redirect=${encodeURIComponent(window.location.pathname)}`,
            );
        }
    }, [user, loading, router]);

    if (!user) return <Loading />;
    if (moderator && user.privileage !== "Moderator")
        return <div>Forbidden</div>;
    return <>{children}</>;
};

export default AuthGuard;
