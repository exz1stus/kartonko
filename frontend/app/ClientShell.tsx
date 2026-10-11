import NavBar from "@/components/Navbar/NavBar";
import ResizableGrid from "@/components/template/ResizableGrid";
import SideBarProvider from "../contexts/SidebarContext";
import HoverProvider from "../contexts/HoverContex";
import { AuthProvider } from "../contexts/AuthContext";
import { Toaster } from "@/components/ui/sonner";
import { AlertDialogProvider } from "@/contexts/AlertDialogContext";
import SideBar from "@/components/SideBar/SideBar";
import Footer from "@/components/Footer";
import GalleryBoardTray from "@/components/Gallery/GalleryBoardTray";
import UploadQueuePersistence from "@/components/UploadImage/UploadQueuePersistence";
import SelectionMarquee from "@/components/Gallery/SelectionMarquee";
import { ThemeProvider } from "next-themes";

const ClientShell = ({ children }: { children: React.ReactNode }) => {
    return (
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
            <AuthProvider>
                <HoverProvider>
                    <SideBarProvider>
                        <AlertDialogProvider>
                            <NavBar />
                            <ResizableGrid sidebar={<SideBar />}>
                                {children}
                            </ResizableGrid>
                            <Footer />
                            <Toaster />
                            <GalleryBoardTray />
                            <SelectionMarquee />
                            <UploadQueuePersistence />
                        </AlertDialogProvider>
                    </SideBarProvider>
                </HoverProvider>
            </AuthProvider>
        </ThemeProvider>
    );
};

export default ClientShell;
