import React from "react"
import ReactMarkdown from "react-markdown"
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'


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
            // Code blocks are rendered with a className of "language-<language>".
            // For example, the following code block:
            //  ```go
            //       fmt.Println("Hello, world!")
            //  ```
            // will be rendered with a className of "language-go",
            // e.g.: <code className="language-go">...</code>
            // matchLanguage will try to match the language from the className.
            const matchLanguage = /language-(\w+)/.exec(className || '')
            // inline is true if the code block is inline, e.g. `fmt.Println("Hello, world!")`
            // If the code is inline, we don't want to render it with SyntaxHighlighter.
            return matchLanguage ? (
              <SyntaxHighlighter
                // eslint-disable-next-line react/no-children-prop
                children={String(children).replace(/\n$/, '')}
                language={matchLanguage[1]}
                PreTag="div"
                {...props}
                style={oneDark}
              />
            ) : (
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
