APP=BingDaily
IDENTIFIER=bingdaily.dacort.github.com

MOD_PATH=$(shell go list -f '{{.Dir}}' github.com/caseymrm/menuet)

include $(MOD_PATH)/menuet.mk