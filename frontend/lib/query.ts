export interface SearchQuery {
    nameContains?: string;
    withTags?: string[];
    userID?: number;
}

export function constructQueryString(query: SearchQuery) {
    const nameQueryString = query.nameContains
        ? `name=${query.nameContains}&`
        : "";
    const tagsQueryString =
        query.withTags && query.withTags.length > 0
            ? `tags=${JSON.stringify(query.withTags)}&`
            : "";
    const userID = query.userID ? `user_id=${query.userID}&` : "";

    let queryString = `${nameQueryString}${tagsQueryString}${userID}`;

    if (queryString.length > 0)
        queryString = queryString.substring(0, queryString.length - 1 - 1);

    return;
}
