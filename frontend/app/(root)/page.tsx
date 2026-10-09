import GalleryServer from "@/components/Gallery/GalleryServer";

export default async function Home({ searchParams }: { searchParams: Promise<{ tag?: string }> }) {
    const { tag } = await searchParams;
    return (
        <>
            <meta
                name="description"
                content="Те, кому не нравятся слова ХУЙ и ПИЗДА, могут идти нахуй. Остальные пруцца!"
            />
            <GalleryServer initialFetchSize={40} initialQuery={{ prefix: "", tags: tag ? [tag] : [] }} />
        </>
    );
}
