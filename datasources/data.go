//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type DatasourceOutput,Config
package datasources

import (
    "errors"
    "fmt"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/ec2"
    "github.com/hashicorp/hcl/v2/hcldec"
    "github.com/hashicorp/packer-plugin-sdk/hcl2helper"
    "github.com/hashicorp/packer-plugin-sdk/template/config"
    "github.com/zclconf/go-cty/cty"
)

type Datasource struct {
    config Config
}

type DatasourceOutput struct {
    VpcId    string `mapstructure:"id"`
    SubnetId string `mapstructure:"subnet_id"`
}

func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
    return d.config.FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Configure(raws ...interface{}) error {
    err := config.Decode(&d.config, nil, raws...)
    if err != nil {
        return err
    }
    return nil
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
    return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

type Config struct {
    VpcTagName string `mapstructure:"vpc_tag_name" required:"true"`
    VpcRegion  string `mapstructure:"vpc_region" required:"true"`
}

func (d *Datasource) Prepare(raws ...interface{}) ([]string, []string, []error) {
    var errs []error
    if d.config.VpcTagName == "" {
        errs = append(errs, errors.New("vpc_tag_name is required"))
    }
    if d.config.VpcRegion == "" {
        errs = append(errs, errors.New("vpc_region is required"))
    }
    if len(errs) > 0 {
        return nil, nil, errs
    }

    return nil, nil, nil
}

func (d *Datasource) Execute() (cty.Value, error) {
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(d.config.VpcRegion),
    })
    if err != nil {
        return cty.NilVal, fmt.Errorf("failued to create AWS session: %s", err)
    }

    svc := ec2.New(sess)

    vpcId, err := getVpc(svc, d)
    if err != nil {
        return cty.NilVal, err
    }

    subnetId, err := getSubnet(svc, vpcId, d)
    if err != nil {
        return cty.NilVal, err
    }

    output := DatasourceOutput{
        VpcId:    aws.StringValue(vpcId),
        SubnetId: aws.StringValue(subnetId),
    }

    return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}

func getVpc(svc *ec2.EC2, d *Datasource) (*string, error) {
    vpcs, err := svc.DescribeVpcs(&ec2.DescribeVpcsInput{
        Filters: []*ec2.Filter{
            {
                Name:   aws.String("tag:Name"),
                Values: []*string{aws.String(d.config.VpcTagName)},
            },
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failued to describe VPCs: %s", err)
    }
    if len(vpcs.Vpcs) == 0 {
        return nil, fmt.Errorf("no VPC found with tag: %s", d.config.VpcTagName)
    }
    vpcId := vpcs.Vpcs[0].VpcId
    return vpcId, nil
}

func getSubnet(svc *ec2.EC2, vpcId *string, d *Datasource) (*string, error) {
    subnets, err := svc.DescribeSubnets(&ec2.DescribeSubnetsInput{
        Filters: []*ec2.Filter{
            {
                Name:   aws.String("vpc-id"),
                Values: []*string{aws.String(*vpcId)},
            },
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failued to describe subnets: %s", err)
    }
    if len(subnets.Subnets) == 0 {
        return nil, fmt.Errorf("no subnets found with vpc: %s", d.config.VpcTagName)
    }
    subnetId := subnets.Subnets[0].SubnetId
    return subnetId, nil
}
