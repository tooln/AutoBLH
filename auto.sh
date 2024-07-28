#!/bin/bash

NEW='\033[1;38;5;154m'
PINK='\033[1;38;5;013m'
BLUE='\033[1;38;5;012m'
GREEN='\033[1;38;5;040m'
CG='\033[1;38;5;087m'
CPO='\033[1;38;5;205m'

start_time=`date +%s`

echo -e ${GREEN}"\n"
figlet -f slant -c "AUTOMATION" | lolcat && figlet -f digital -c "Version: 1.0" | lolcat

echo -n -e ${CG} "\n\n\n[+] Enter your Domain: "
read domain

#Subdomain Finder
echo -e ${NEW}"\n\nSubdomain Finder has started...                       Please wait a little bit...\n\n"
echo -e ${PINK}"\n"
subfinder -d $domain -o /root/sub/$domain.txt -all
sleep 1

#Subdomain Takeover
echo -e ${NEW}"\n\nSubdomain Takeover Scanner has started...                       Please wait a little bit...\n\n"
echo -e ${GREEN}"\n"
subzy run --targets /root/sub/$domain.txt --hide_fails > $domain.subzy.txt
echo -e ${CPO}"\n\n                  Scanning has finished.\n\n"
sleep 1

#HTTPX
echo -e ${NEW}"\n\nHTTP Probing has started...                       Please wait a little bit...\n\n"
sleep 1
echo -e ${PINK}"\n"
cat /root/sub/$domain.txt | httpx > /root/subx/$domain.httpx.txt
sleep 1

#HTTPut File Uploading
echo -e ${NEW}"\n\nFile Uploading has started...                       Please wait a little bit...\n\n"
sleep 1
echo -e ${PINK}"\n"
go run httput.go -file=/root/subx/$domain.httpx.txt -output=$domain.httput.txt
sleep 1

#AutoBLH
echo -e ${NEW}"\n\nAutoBLH has started...                       Please wait a little bit...\n\n"
sleep 1
echo -e ${PINK}"\n"
gau $domain --subs --o $domain.gauLinks.txt &&
sort -u $domain.gauLinks.txt -o $domain.links.txt &&
go run link_checker.go -links=$domain.links.txt -output=$domain.aliveLinks.txt &&
go run SMLFinder.go -links=$domain.aliveLinks.txt -output=$domain.output.txt -save=$domain.savefile.txt &&
sed 's/"//g' $domain.output.txt | sort | uniq &&
sleep 1

#Nuclei Scanner
echo -e ${CPO}"\nNuclei Templates Critical Scanning Started:\n\n"
echo -e ${CG}"\n"
nuclei -l /root/subx/$domain.httpx.txt -s critical -retries 3 -o $domain.nuclei.critical.txt
echo -e ${CPO}"\nNuclei Templates High Scanning Started:\n\n"
echo -e ${BLUE}"\n"
nuclei -l /root/subx/$domain.httpx.txt -s high -retries 3 -o $domain.nuclei.high.txt
echo -e ${CPO}"\nNuclei Templates Medium Scanning Started:\n\n"
echo -e ${GREEN}"\n"
nuclei -l /root/subx/$domain.httpx.txt -s medium -retries 3 -o $domain.nuclei.medium.txt
echo -e ${CPO}"\n\n\nNuclei Templates Low Scanning Started:\n\n"
echo -e ${PINK}"\n"
nuclei -l /root/subx/$domain.httpx.txt -s low -retries 3 -o $domain.nuclei.low.txt
echo -e ${CPO}"\n\n                  Scanning has finished.\n\n"
sleep 1

end_time=`date +%s`
echo -e ${CPO}"\n\nExecution time was `expr $end_time - $start_time` second.\n\n"
