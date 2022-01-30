#image/quickfeed:go

printf "Creator Access Token:        {{ .CreatorAccessToken }}\n"
printf "Assignment Name:             {{ .AssignmentName }}"
printf "Student Repository URL:      {{ .GetURL }}"
printf "Course Tests Repository URL: {{ .TestURL }}"
printf "QuickFeed Session Secret:    {{ .RandomSecret }}"
