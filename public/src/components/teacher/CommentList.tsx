import * as React from "react";
import { Comment, User, CommentWithUser } from '../../../proto/ag_pb';
import { IComment } from "./Comment";

interface CommentListProps {
    comments: CommentWithUser[];
    commenting: boolean;
    updateComment: (comment: Comment) => void;
    deleteComment: (commentID: number) => void;
    toggleCommenting: (on: boolean) => void;
}

interface CommentListState {
    selectedComment?: CommentWithUser,
}

export class CommentList extends React.Component<CommentListProps> {

    public render() {
        return <div><div className="row comment-list list-group">
            {
                this.props.comments.map((c, i) => <IComment
                    key={"cm" + i}
                    author={c.getUser()}
                    comment={c.getComment()}
                    onSelect={() => this.setState({
                        selectedComment: c,
                    })}
                />)
            }
        </div>
        <div className="row comment-add">

        </div>
        </div>
    }
    // Text: add a new comment (button+icon)
    // - input for a new comment
    // switching adding/add
    // - edit a comment
}