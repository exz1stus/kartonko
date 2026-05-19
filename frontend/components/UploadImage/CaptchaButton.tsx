import { cn } from "@/lib/utils";
import { Turnstile } from "@marsidev/react-turnstile";
import React, { useState } from "react";

interface Props extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    onVerifySuccess: (token: string) => void;
    verifyingText?: string;
}

const CaptchaButton = ({
    children,
    verifyingText = "Verifying...",
    className,
    type,
    disabled,
    onVerifySuccess,
    ...props
}: Props) => {
    const [captchaToken, setCaptchaToken] = useState<string | null>(null);
    return (
        <div className="flex flex-col items-center gap-4">
            <Turnstile
                className={`${captchaToken && "hidden"} flex justify-center`}
                siteKey={process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY!}
                onSuccess={(token) => {
                    setCaptchaToken(token);
                    onVerifySuccess(token);
                }}
                onExpire={() => setCaptchaToken(null)}
                onError={() => setCaptchaToken(null)}
            />
            {captchaToken && (
                <button
                    {...props}
                    className={cn(
                        "px-12 py-1 border border-surface-20 hover:border-surface-30 rounded-2xl w-96 transition cursor-pointer glass",
                        className,
                    )}
                    type={type}
                    disabled={disabled || !captchaToken}
                >
                    {children}
                </button>
            )}
        </div>
    );
};

export default CaptchaButton;
