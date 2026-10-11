export type UploadAction = {
    kind: "add_to_board";
    boardId: number;
    boardName: string;
};

export interface UploadIntent {
    tags?: string[];
    actions?: UploadAction[];
}

export const uploadActionKey = (action: UploadAction) =>
    `${action.kind}:${action.boardId}`;

export const mergeUploadActions = (first: UploadAction[], second: UploadAction[]) =>
    [...new Map([...first, ...second].map((action) => [uploadActionKey(action), action])).values()];
