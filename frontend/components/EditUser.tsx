"use client";
import { Pencil, Save, Trash2, X } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { deleteImagesByQuery, patchUser } from "@/lib/api/generated/client";
import { UserDataResponse } from "@/lib/api/generated/model/userDataResponse";
import { sanitizeName } from "@/lib/sanitizeName";
import ImageSelector from "@/components/ImageSelector";
import { useTranslations } from "@/contexts/LocaleContext";

interface Props {
    user: UserDataResponse;
    hasPermission: boolean;
}

const EditUser = ({ user, hasPermission }: Props) => {
    const { t } = useTranslations();
    const router = useRouter();
    const [editing, setEditing] = useState(false);
    const [username, setUsername] = useState(user.username);
    const [pending, setPending] = useState(false);
    const [pictureSource, setPictureSource] = useState<File | null>(null);

    useEffect(() => {
        if (editing) return;
        setUsername(user.username);
    }, [editing, user.username]);

    const save = useCallback(async () => {
        if (pending) return;
        setPending(true);
        try {
            const nextUsername = sanitizeName(username.trim());
            if (!nextUsername) throw new Error(t("profile.usernameEmpty"));

            const updated = await patchUser(
                user.id,
                { username: nextUsername },
                { credentials: "include" },
            );
            setEditing(false);
            router.replace(`/user/${encodeURIComponent(updated.username)}`);
            router.refresh();
        } finally {
            setPending(false);
        }
    }, [pending, router, t, user.id, username]);

    const uploadPicture = useCallback(
        async (file: File) => {
            if (pending) return;
            setPending(true);
            try {
                const form = new FormData();
                form.append("picture", file);
                const response = await fetch(
                    `/apilocal/user/${user.id}/picture`,
                    {
                        method: "POST",
                        credentials: "include",
                        body: form,
                    },
                );
                const result = await response.json().catch(() => null);
                if (!response.ok)
                    throw new Error(
                        result?.error ?? t("error.couldNotUploadProfilePicture"),
                    );
                router.refresh();
            } finally {
                setPending(false);
            }
        },
        [pending, router, t, user.id],
    );

    const onPictureSelected = async (file: File, _mode: "fit" | "crop", sourceFile: File) => {
        const upload = uploadPicture(file);
        await toast.promise(upload, {
            loading: t("profile.pictureUploading"),
            success: t("profile.pictureUpdated"),
            error: (error) => error.message,
        });
        setPictureSource(sourceFile);
    };

    const deleteUsersImages = useCallback(async () => {
        if (pending) return;
        if (!window.confirm(t("profile.deleteImagesConfirm", { name: user.username })))
            return;
        setPending(true);
        toast.promise(
            deleteImagesByQuery(
                { user_id: user.id },
                { credentials: "include" },
            ).finally(() => setPending(false)),
            {
                loading: t("common.loading"),
                success: () => {
                    router.refresh();
                    return t("profile.deleteImages");
                },
                error: (error) => error.message,
            },
        );
    }, [pending, router, t, user.id, user.username]);

    if (!hasPermission) return null;

    return (
        <div className="flex flex-col items-center gap-2">
            {hasPermission && (
                <>
                    {!editing && (
                        <button
                            type="button"
                            onClick={() => setEditing(true)}
                            className="flex flex-row items-center gap-2 px-2 py-1 border hover:border-primary rounded-2xl"
                        >
                            <Pencil size={16} /> {t("profile.edit")}
                        </button>
                    )}
                    {editing && (
                        <div className="flex flex-col gap-2 w-full max-w-sm">
                            <ImageSelector
                                label={t("settings.background.browse")}
                                editLabel={t("imageSelector.adjustSelected")}
                                aspectRatio={1}
                                currentFile={pictureSource}
                                accept="image/png,image/jpeg,image/gif"
                                outputType="image/jpeg"
                                maxOutputDimension={1024}
                                disabled={pending}
                                onSelect={onPictureSelected}
                            />
                            <input
                                className="field-fill px-2 py-1"
                                value={username}
                                onChange={(event) =>
                                    setUsername(
                                        sanitizeName(event.target.value),
                                    )
                                }
                                aria-label={t("profile.username")}
                            />
                            <div className="flex gap-2">
                                <button
                                    type="button"
                                    disabled={pending}
                                    onClick={() =>
                                        void toast.promise(save(), {
                                            loading: t("image.saving"),
                                            success: t("profile.updated"),
                                            error: (error) => error.message,
                                        })
                                    }
                                    className="inline-flex items-center gap-1 px-2 py-1 border hover:border-primary rounded-md"
                                >
                                    <Save size={16} /> {t("common.save")}
                                </button>
                                <button
                                    type="button"
                                    disabled={pending}
                                    onClick={() => {
                                        setUsername(user.username);
                                        setEditing(false);
                                    }}
                                    className="inline-flex items-center gap-1 px-2 py-1 border rounded-md"
                                >
                                    <X size={16} /> {t("common.cancel")}
                                </button>
                            </div>
                            <button
                                type="button"
                                disabled={pending}
                                onClick={() => void deleteUsersImages()}
                                className="flex flex-row items-center gap-2 px-2 py-1 border hover:border-red-500 rounded-2xl hover:text-red-500"
                            >
                                <Trash2 size={16} />
                                <span>{t("profile.deleteImages")}</span>
                            </button>
                        </div>
                    )}
                </>
            )}
        </div>
    );
};

export default EditUser;
