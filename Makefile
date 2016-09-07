.PHONY: start-local-gitlab start-local-gitlab-runner coverage coverage-html

start-local-gitlab:
	gcloud compute firewall-rules create allow-http-for-http-tag --allow tcp:80 --target-tags http
	gcloud compute firewall-rules create allow-https-for-https-tag --allow tcp:443 --target-tags https
	gcloud compute instances start gitlab || gcloud compute instances create gitlab --boot-disk-size 10GB --machine-type n1-standard-2 --tags http,https --image-family=ubuntu-1404-lts --image-project=ubuntu-os-cloud
	~/dev/domains_and_tls/create_route53_A_record_to_ip_address.sh danstutzman.com gitlab.danstutzman.com `~/dev/domains_and_tls/ip_address_for_gcloud.sh gitlab`
	gcloud compute ssh gitlab -C 'echo sudo shutdown +120 -P | at now; \
    sudo debconf-set-selections \
			<<< "postfix postfix/mailname string your.hostname.com"; \
    sudo debconf-set-selections \
			<<< "postfix postfix/main_mailer_type string \"Internet Site\""; \
    sudo apt-get install -y curl openssh-server ca-certificates postfix mailutils; \
		curl -sS https://packages.gitlab.com/install/repositories/gitlab/gitlab-ce/script.deb.sh | sudo bash; \
		sudo apt-get install -y gitlab-ce; \
	  sudo gitlab-ctl reconfigure; \
		'

start-local-gitlab-runner:
	gcloud compute ssh gitlab -C 'curl -L https://packages.gitlab.com/install/repositories/runner/gitlab-ci-multi-runner/script.deb.sh | sudo bash; \
    sudo apt-get install -y gitlab-ci-multi-runner; \
		echo "Answers to questions: 1) http://gitlab.danstutzman.com/ci 2) See http://gitlab.danstutzman.com/root/todomvc-backend-go/runners for runners_token 3) (blank) 4) (blank) 5) shell" \
		sudo gitlab-ci-multi-runner register; \
		'

coverage:
	echo "mode: count" > .coverage-all.out
	for PACKAGE in . ./model ./web; do \
		echo > .coverage.out; \
		go test -coverprofile=.coverage.out -covermode=count $$PACKAGE; \
		tail -n +2 .coverage.out >> .coverage-all.out; \
	done
	rm .coverage.out
	cat .coverage-all.out | awk -F' ' '{ all_lines += $$2; covered_lines += $$3 } END { print 100 * covered_lines / all_lines, "% covered" }'

coverage-html: coverage
	go tool cover -html=.coverage-all.out
