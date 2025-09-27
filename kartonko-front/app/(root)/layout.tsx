import "@/app/globals.css";
import { AuthProvider } from "../AuthContext";
import NavBar from "@/components/Navbar/NavBar";
import SideBarProvider from "../SidebarProvider";
import ResizableGrid from "@/components/ResizableGrid";
import { ScrollArea } from "@/components/ui/scroll-area";

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {


    return (
        <html lang="en" className="h-screen">
            <AuthProvider>
                <SideBarProvider>
                    <body className="h-full grid grid-rows-[auto_1fr]">
                        <NavBar />
                        <ResizableGrid>
                            <main className="h-full flex flex-col">
                                {children}
                            </main>
                        </ResizableGrid>
                    </body>
                </SideBarProvider>
            </AuthProvider>
        </html >
    );
}


