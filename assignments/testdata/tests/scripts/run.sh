#image/quickfeed:go

printf "Creator Access Token:        {{ .CreatorAccessToken }}\n"
printf "Assignment Name:             {{ .AssignmentName }}\n"
printf "Student Repository URL:      {{ .GetURL }}\n"
printf "Course Tests Repository URL: {{ .TestURL }}\n"
printf "QuickFeed Session Secret:    {{ .RandomSecret }}\n"
