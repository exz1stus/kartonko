import NavBar from "@/components/Navbar/NavBar";
import ResizableGrid from "@/components/template/ResizableGrid";
import SideBarProvider from "../contexts/SidebarContext";
import HoverProvider from "../contexts/HoverContex";
import { AuthProvider } from "../contexts/AuthContext";
import { Toaster } from "@/components/ui/sonner";
import { AlertDialogProvider } from "@/contexts/AlertDialogContext";

const ClientShell = ({ children }: { children: React.ReactNode }) => {
    return (
        <AuthProvider>
            <HoverProvider>
                <SideBarProvider>
                    <AlertDialogProvider>
                        <body
                            className="grid grid-rows-[auto_1fr] bg-image bg-surface-10 h-screen overflow-hidden"
                            style={{
                                backgroundImage: `url(${process.env.NEXT_PUBLIC_API_ORIGIN}/image/raw/bg)`,
                            }}
                        >
                            <NavBar />
                            <ResizableGrid>{children}</ResizableGrid>
                            <Toaster />
                        </body>
                    </AlertDialogProvider>
                </SideBarProvider>
            </HoverProvider>
        </AuthProvider>
    );
};

export default ClientShell;
