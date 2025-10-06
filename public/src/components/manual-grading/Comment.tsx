import React from "react"
import ReactMarkdown from "react-markdown"


const CriterionComment = ({ comment }: { comment: string }) => {
  if (comment == "" || comment.length == 0) {
    return null
  }

  return (
    <div className="comment-md">
      <ReactMarkdown
        // eslint-disable-next-line react/no-children-prop
        children={comment}
        components={{
          // eslint-disable-next-line @typescript-eslint/no-unused-vars
          code({ node, className, children, ref, ...props }) {
            // TODO: Add `react-syntax-highlighter` to highlight code blocks
            // TODO: Syntax highlighting was removed pending a fix for a vulnerability
            // Code blocks are rendered with a className of "language-<language>".
            // For example, the following code block:
            //  ```go
            //       fmt.Println("Hello, world!")
            //  ```
            // will be rendered with a className of "language-go",
            // e.g.: <code className="language-go">...</code>
            // matchLanguage will try to match the language from the className.
            // inline is true if the code block is inline, e.g. `fmt.Println("Hello, world!")`
            // If the code is inline, we don't want to render it with SyntaxHighlighter.
            return (
              <code className={className} {...props}>
                {children}
              </code>
            )
          }
        }}
      />
    </div>
  )
}

export default CriterionComment
