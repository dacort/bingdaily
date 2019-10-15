APP=BingDaily
REPO=dacort/bingdaily
IDENTIFIER=bingdaily.dacort.github.com
IDENTITY=Developer ID Application: Damon Cortesi (CR35Z2ZUAT)

MOD_PATH=$(shell go list -f '{{.Dir}}' github.com/caseymrm/menuet)

include $(MOD_PATH)/menuet.mk
