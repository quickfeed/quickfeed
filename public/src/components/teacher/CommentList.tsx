import * as React from "react";
import { Comment, User, CommentWithUser } from "../../../proto/ag_pb";
import { IComment } from "./Comment";

interface CommentListProps {
    comments: CommentWithUser[];
}

export class CommentList extends React.Component<CommentListProps> {

    public render() {
        return <div className="row comment-list list-group">
            {
                this.props.comments.map((c, i) => <IComment
                    key={"cm" + i}
                    author={c.getUser()}
                    comment={c.getComment()}
                />)
            }
        </div>
    }
}