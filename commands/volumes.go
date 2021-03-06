/*
Copyright 2016 The Doctl Authors All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commands

import (
	"fmt"

	"github.com/digitalocean/doctl"
	"github.com/digitalocean/doctl/do"
	"github.com/digitalocean/godo"
	"github.com/dustin/go-humanize"
	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
)

// Volume creates the Volume command
func Volume() *Command {
	cmd := &Command{
		Command: &cobra.Command{
			Use:   "volume",
			Short: "volume commands",
			Long:  "volume is used to access volume commands",
		},
	}

	cmdRunVolumeList := CmdBuilder(cmd, RunVolumeList, "list", "list volume", Writer,
		aliasOpt("ls"), displayerType(&volume{}))
	AddStringFlag(cmdRunVolumeList, doctl.ArgRegionSlug, "", "", "Volume region")

	cmdVolumeCreate := CmdBuilder(cmd, RunVolumeCreate, "create <volume-name>", "create a volume", Writer,
		aliasOpt("c"), displayerType(&volume{}))
	AddStringFlag(cmdVolumeCreate, doctl.ArgVolumeSize, "", "4TiB", "Volume size",
		requiredOpt())
	AddStringFlag(cmdVolumeCreate, doctl.ArgVolumeDesc, "", "", "Volume description")
	AddStringFlag(cmdVolumeCreate, doctl.ArgVolumeRegion, "", "", "Volume region",
		requiredOpt())

	CmdBuilder(cmd, RunVolumeDelete, "delete <volume-id>", "delete a volume", Writer,
		aliasOpt("rm"))

	CmdBuilder(cmd, RunVolumeGet, "get <volume-id>", "get a volume", Writer, aliasOpt("g"),
		displayerType(&volume{}))

	cmdRunVolumeSnapshot := CmdBuilder(cmd, RunVolumeSnapshot, "snapshot <volume-id>", "create a volume snapshot", Writer,
		aliasOpt("s"), displayerType(&volume{}))
	AddStringFlag(cmdRunVolumeSnapshot, doctl.ArgSnapshotName, "", "", "Snapshot name", requiredOpt())
	AddStringFlag(cmdRunVolumeSnapshot, doctl.ArgSnapshotDesc, "", "", "Snapshot description")

	return cmd

}

// RunVolumeList returns a list of volumes.
func RunVolumeList(c *CmdConfig) error {

	al := c.Volumes()

	region, err := c.Doit.GetString(c.NS, doctl.ArgRegionSlug)
	if err != nil {
		return nil
	}

	matches := []glob.Glob{}
	for _, globStr := range c.Args {
		g, err := glob.Compile(globStr)
		if err != nil {
			return fmt.Errorf("unknown glob %q", globStr)
		}

		matches = append(matches, g)
	}

	list, err := al.List()
	if err != nil {
		return err
	}
	var matchedList []do.Volume

	for _, volume := range list {
		var skip = true
		if len(matches) == 0 {
			skip = false
		} else {
			for _, m := range matches {
				if m.Match(volume.ID) {
					skip = false
				}
				if m.Match(volume.Name) {
					skip = false
				}
			}
		}

		if !skip && region != "" {
			if region != volume.Region.Slug {
				skip = true
			}
		}

		if !skip {
			matchedList = append(matchedList, volume)
		}
	}
	item := &volume{volumes: matchedList}
	return c.Display(item)
}

// RunVolumeCreate creates a volume.
func RunVolumeCreate(c *CmdConfig) error {
	if len(c.Args) == 0 {
		return doctl.NewMissingArgsErr(c.NS)
	}

	name := c.Args[0]

	sizeStr, err := c.Doit.GetString(c.NS, doctl.ArgVolumeSize)
	if err != nil {
		return err
	}
	size, err := humanize.ParseBytes(sizeStr)
	if err != nil {
		return err
	}

	desc, err := c.Doit.GetString(c.NS, doctl.ArgVolumeDesc)
	if err != nil {
		return err
	}

	region, err := c.Doit.GetString(c.NS, doctl.ArgVolumeRegion)
	if err != nil {
		return err

	}

	var createVolume godo.VolumeCreateRequest

	createVolume.Name = name
	createVolume.SizeGigaBytes = int64(size / (1 << 30))
	createVolume.Description = desc
	createVolume.Region = region

	al := c.Volumes()

	d, err := al.CreateVolume(&createVolume)
	if err != nil {
		return err
	}
	item := &volume{volumes: []do.Volume{*d}}
	return c.Display(item)

}

// RunVolumeDelete deletes a volume.
func RunVolumeDelete(c *CmdConfig) error {
	if len(c.Args) == 0 {
		return doctl.NewMissingArgsErr(c.NS)

	}
	id := c.Args[0]
	al := c.Volumes()
	if err := al.DeleteVolume(id); err != nil {
		return err

	}
	return nil
}

// RunVolumeGet gets a volume.
func RunVolumeGet(c *CmdConfig) error {
	if len(c.Args) == 0 {
		return doctl.NewMissingArgsErr(c.NS)

	}
	id := c.Args[0]
	al := c.Volumes()
	d, err := al.Get(id)
	if err != nil {
		return err
	}
	item := &volume{volumes: []do.Volume{*d}}
	return c.Display(item)
}

// RunVolumeSnapshot creates a snapshot of a volume
func RunVolumeSnapshot(c *CmdConfig) error {
	var err error
	if len(c.Args) == 0 {
		return doctl.NewMissingArgsErr(c.NS)
	}

	al := c.Volumes()
	id := c.Args[0]

	name, err := c.Doit.GetString(c.NS, doctl.ArgSnapshotName)
	if err != nil {
		return err
	}

	desc, err := c.Doit.GetString(c.NS, doctl.ArgSnapshotDesc)
	if err != nil {
		return nil
	}

	req := &godo.SnapshotCreateRequest{
		VolumeID:    id,
		Name:        name,
		Description: desc,
	}

	if _, err := al.CreateSnapshot(req); err != nil {
		return err
	}

	return nil
}
