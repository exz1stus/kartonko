import NavBar from "@/components/Navbar/NavBar";
import ResizableGrid from "@/components/template/ResizableGrid";
import SideBarProvider from "../contexts/SidebarContext";
import HoverProvider from "../contexts/HoverContex";
import { AuthProvider } from "../contexts/AuthContext";
import { Toaster } from "@/components/ui/sonner";
import { AlertDialogProvider } from "@/contexts/AlertDialogContext";
import SideBar from "@/components/SideBar/SideBar";
import Footer from "@/components/Footer";

const ClientShell = ({ children }: { children: React.ReactNode }) => {
    return (
        <AuthProvider>
            <HoverProvider>
                <SideBarProvider>
                    <AlertDialogProvider>
                        <body
                            className="grid grid-rows-[auto_1fr] bg-image bg-surface-10 h-screen overflow-hidden"
                            //TODO: Add bg image
                            // style={{
                            //     backgroundImage: `url(${process.env.NEXT_PUBLIC_API_ORIGIN}/image/raw/bg)`,
                            // }}
                        >
                            <NavBar />
                            <ResizableGrid sidebar={<SideBar />}>
                                {children}
                            </ResizableGrid>
                            <Footer />
                            <Toaster />
                        </body>
                    </AlertDialogProvider>
                </SideBarProvider>
            </HoverProvider>
        </AuthProvider>
    );
};

export default ClientShell;
