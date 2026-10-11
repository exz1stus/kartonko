"use client";

import Image from "next/image";
import { useState } from "react";
import { useTranslations } from "@/contexts/LocaleContext";

const fallbackPicture =
    "https://cdn-icons-png.flaticon.com/512/149/149071.png";

export default function ProfileImage({
    src,
    size,
    className,
}: {
    src?: string | null;
    size: number;
    className?: string;
}) {
    const pictureSrc = src || fallbackPicture;
    return (
        <ProfileImageSource
            key={pictureSrc}
            src={pictureSrc}
            size={size}
            className={className}
        />
    );
}

function ProfileImageSource({
    src,
    size,
    className,
}: {
    src: string;
    size: number;
    className?: string;
}) {
    const [currentSrc, setCurrentSrc] = useState(src);
    const { t } = useTranslations();

    return (
        <Image
            src={currentSrc}
            alt={t("profile.pictureAlt")}
            width={size}
            height={size}
            className={className}
            unoptimized
            onError={() => {
                if (currentSrc !== fallbackPicture) setCurrentSrc(fallbackPicture);
            }}
        />
    );
}
