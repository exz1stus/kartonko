"use client";

import * as React from "react";
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";

type DialogOptions = {
    title?: string;
    confirmText?: string;
    cancelText?: string;
};

type DialogContextType = {
    openDialog: (message: string, options?: DialogOptions) => Promise<boolean>;
};

const DialogContext = React.createContext<DialogContextType | undefined>(
    undefined,
);

export const AlertDialogProvider: React.FC<{ children: React.ReactNode }> = ({
    children,
}) => {
    const [state, setState] = React.useState<{
        open: boolean;
        message: string;
        options: DialogOptions;
        resolve?: (value: boolean) => void;
    }>({ open: false, message: "", options: {} });

    const openDialog = (message: string, options: DialogOptions = {}) => {
        return new Promise<boolean>((resolve) => {
            setState({ open: true, message, options, resolve });
        });
    };

    const handleClose = (result: boolean) => {
        if (state.resolve) state.resolve(result);
        setState({ ...state, open: false, resolve: undefined });
    };

    return (
        <DialogContext.Provider value={{ openDialog }}>
            {children}
            <AlertDialog
                open={state.open}
                onOpenChange={() => handleClose(false)}
            >
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>
                            {state.options.title || "Confirm"}
                        </AlertDialogTitle>
                        <AlertDialogDescription>
                            {state.message}
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter className="flex gap-2">
                        <AlertDialogCancel onClick={() => handleClose(false)}>
                            {state.options.cancelText || "Cancel"}
                        </AlertDialogCancel>
                        <AlertDialogAction onClick={() => handleClose(true)}>
                            {state.options.confirmText || "OK"}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </DialogContext.Provider>
    );
};

export const useDialog = () => {
    const context = React.useContext(DialogContext);
    if (!context)
        throw new Error("useDialog must be used within an AlertDialogProvider");
    return context.openDialog;
};
