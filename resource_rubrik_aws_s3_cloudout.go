package main

import (
	"errors"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/rubrikinc/rubrik-sdk-for-go/rubrikcdm"
)

func resourceRubrikAWSS3CloudOut() *schema.Resource {
	return &schema.Resource{
		Create: resourceRubrikAWSS3CloudOutCreate,
		Read:   resourceRubrikAWSS3CloudOutRead,
		Update: resourceRubrikAWSS3CloudOutUpdate,
		Delete: resourceRubrikAWSS3CloudOutDelete,

		Schema: map[string]*schema.Schema{
			"aws_bucket": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the AWS S3 bucket you wish to use as an archive target.",
			},
			"storage_class": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"standard",
					"standard_ia",
					"reduced_redundancy",
				}, true),
				Default:     "standard",
				Description: "The AWS storage class you wish to use.",
			},
			"archive_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the archive location used in the Rubrik GUI.",
			},
			"aws_region": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"ap-south-1",
					"ap-northeast-3",
					"ap-northeast-2",
					"ap-southeast-1",
					"ap-southeast-2",
					"ap-northeast-1",
					"ca-central-1",
					"cn-north-1",
					"cn-northwest-1",
					"eu-central-1",
					"eu-west-1",
					"eu-west-2",
					"eu-west-3",
					"us-west-1",
					"us-east-1",
					"us-east-2",
					"us-west-2",
				}, true),
			},
			"aws_access_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The access key of a AWS account with the required permissions.",
			},
			"aws_secret_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The secret key of a AWS account with the required permissions.",
				Sensitive:   true,
			},
			"rsa_key": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"kms_master_key_id"},
				Description:   "The RSA key that will be used to encrypt the archive data.",
			},
			"kms_master_key_id": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"rsa_key"},
				Description:   "The AWS KMS master key ID that will be used to encrypt the archive data.",
			},
			"timeout": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     15,
				Description: "The number of seconds to wait to establish a connection the Rubrik cluster before returning a timeout error.",
			},
		},
	}

}

func resourceRubrikAWSS3CloudOutCreate(d *schema.ResourceData, meta interface{}) error {

	_, rsaOk := d.GetOk("rsa_key")
	_, kmsOk := d.GetOk("kms_master_key_id")

	if !rsaOk && !kmsOk {
		return errors.New("Either `rsa_key` or `kms_master_key_id` must be provided")
	}

	rubrik := meta.(*rubrikcdm.Credentials)

	log.Println("[INFO] Creating the S3 archival location")
	_, err := rubrik.AWSS3CloudOutRSA(d.Get("aws_bucket").(string), d.Get("storage_class").(string), d.Get("archive_name").(string), d.Get("aws_region").(string), d.Get("aws_access_key").(string), d.Get("aws_secret_key").(string), d.Get("rsa_key").(string), d.Get("timeout").(int))

	if err != nil {
		return err
	}

	d.SetId(d.Get("aws_bucket").(string))

	return resourceRubrikAWSS3CloudOutRead(d, meta)
}

func resourceRubrikAWSS3CloudOutRead(d *schema.ResourceData, meta interface{}) error {

	rubrik := meta.(*rubrikcdm.Credentials)

	log.Println("[INFO] Searching the Rubrik cluster for the current archival locations.")
	archivesOnCluster, err := rubrik.Get("internal", "/archive/object_store")
	if err != nil {
		return err
	}

	if archivesOnCluster.(map[string]interface{})["total"].(float64) == 0 {
		d.SetId("")
	}

	log.Println("[INFO] Searching the Rubrik cluster for the current archival locations ")
	var archivePresent = true
	for _, v := range archivesOnCluster.(map[string]interface{})["data"].([]interface{}) {
		archiveDefinition := (v.(interface{}).(map[string]interface{})["definition"]).(interface{}).(map[string]interface{})

		if archiveDefinition["objectStoreType"] == "S3" && archiveDefinition["name"] == d.Get("aws_bucket").(string) {

			d.Set("archive_name", archiveDefinition["name"])
			d.Set("aws_bucket", archiveDefinition["bucket"])
			d.Set("storage_class", archiveDefinition["storageClass"])
			d.Set("aws_region", archiveDefinition["defaultRegion"])
			d.Set("aws_access_key", archiveDefinition["accessKey"])

			archivePresent = true
		}
	}

	if archivePresent == false {
		d.SetId("")
	}

	return nil

}

func resourceRubrikAWSS3CloudOutUpdate(d *schema.ResourceData, meta interface{}) error {

	// rubrik := meta.(*rubrikcdm.Credentials)

	return nil
}

func resourceRubrikAWSS3CloudOutDelete(d *schema.ResourceData, m interface{}) error {
	// Cluster Timezone is a requirement for the Rubrik cluster and can not be "deleted"
	return nil
}