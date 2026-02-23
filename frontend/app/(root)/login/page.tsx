"use client";
import { useAuth } from "@/contexts/AuthContext";
import { useSearchParams } from "next/navigation";
import { useEffect } from "react";

export default function LoginPage() {
    const { user, login, loading } = useAuth();

    const searchParams = useSearchParams();

    const redirectTo = searchParams.get("redirect");

    useEffect(() => {
        if (!loading && !user) {
            login(redirectTo || undefined);
        }
    }, [user, loading, login]);

    if (loading) {
        return <div>Loading...</div>;
    }

    if (user) {
        return <div>You are already logged in as {user.username}</div>;
    }

    return <div>Redirecting to login...</div>;
}
