import ImageMetadata from "@/lib/image";
import Gallery from "./Gallery";
import { SearchQuery } from "./ImageSearch";
import { serverFetch } from "@/lib/serverFetch";
import { constructQueryString } from "@/lib/query";

interface Props {
    initialFetchSize: number;
    initialQuery?: SearchQuery;
}

const INITIAL_QUERY: SearchQuery = {
    nameContains: "",
    withTags: [],
};

const GalleryServer = async ({
    initialFetchSize,
    initialQuery = INITIAL_QUERY,
}: Props) => {
    const fetchImages = async (
        intialFetchSize: number,
        initialQuery: SearchQuery,
    ) => {
        const queryString = constructQueryString(initialQuery);
        const response = await serverFetch(
            `/images?${queryString}&limit=${intialFetchSize}`,
            {
                cache: "no-store",
            },
        );

        if (!response.ok) throw new Error("Failed to fetch images");
        console.log(
            `server intial fetching cursor ${0} limit ${intialFetchSize}`,
        );
        const data: ImageMetadata[] = await response.json();
        return data;
    };

    let images: ImageMetadata[] = [];
    let initReachedEnd = false;
    if (initialFetchSize > 0) {
        try {
            images = await fetchImages(initialFetchSize, initialQuery);
            initReachedEnd = images.length < initialFetchSize;
        } catch (error) {
            console.error("Error fetching images:", error);
            images = [];
        }
    }

    return (
        <Gallery
            initialImages={images}
            initReachedEnd={initReachedEnd}
            initialQuery={initialQuery}
        />
    );
};

export default GalleryServer;
