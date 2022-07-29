package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

const (
	FontStyleNormal     string = ""
	FontStyleBold       string = "B"
	FontStyleItalic     string = "I"
	FontStyleBoldItalic string = "BI"
)

var (
	WorkingPageWidth float64
	DefaultFont      string
)

func init() {
	os.Setenv("TZ", "UTC")

	initFlags()
}

func main() {
	parseFlags()
	c := parseConfiguration()

	pdf := pdfGlobal(c)

	pdf.AddPage()

	pdfContactLine(pdf, c)
	pdfSkillsSection(pdf, &c.Skills, &c.Controls.Skills.First)
	pdfSkillsSection(pdf, &c.Skills, &c.Controls.Skills.Second)
	pdfOrganizationalExperience(pdf, &c.Employment, &c.Controls.Employers, false)
	pdfOrganizationalExperience(pdf, &c.Employment, &c.Controls.Employers, true)
	pdfOrganizationalExperience(pdf, &c.Politics, &c.Controls.Politics, false)
	pdfOrganizationalExperience(pdf, &c.Politics, &c.Controls.Politics, true)
	pdfOrganizationalExperience(pdf, &c.Volunteering, &c.Controls.Volunteering, false)
	pdfOrganizationalExperience(pdf, &c.Volunteering, &c.Controls.Volunteering, true)
	pdfSkillsSection(pdf, &c.Skills, &c.Controls.Skills.Third)
	pdfEducation(pdf, c)
	pdfProjects(pdf, c)
	pdfCertifications(pdf, c)

	writePdf(pdf, c)
}

func pdfGlobal(c *Configuration) *gofpdf.Fpdf {
	titleSubject := c.Contact.Name + "'s Resume"

	timeDate := time.Now()

	pdf := gofpdf.New(gofpdf.OrientationPortrait, gofpdf.UnitMillimeter, gofpdf.PageSizeLetter, "")

	pdf.SetTitle(titleSubject, false)
	pdf.SetSubject(titleSubject, false)
	pdf.SetAuthor(c.Contact.Name, false)
	pdf.SetCreator(c.Contact.Name, false)
	pdf.SetKeywords(strings.Join(c.Controls.Pdf.Keywords, " "), false)
	pdf.SetDisplayMode("fullwidth", "SinglePage")
	// pdf.SetProtection(gofpdf.CnProtectPrint, "", "")
	pdf.SetCreationDate(timeDate)
	pdf.SetModificationDate(timeDate)

	pdf.SetMargins(c.Controls.Pdf.Margins.Left, c.Controls.Pdf.Margins.Top, c.Controls.Pdf.Margins.Right)

	pdf.SetHeaderFunc(func() {
		pdf.SetFont(c.Controls.Pdf.Fonts.Header, FontStyleBoldItalic, 18)

		pdf.Cell(0, 0, c.Contact.Name)

		pdf.SetFontSize(14)

		pdf.CellFormat(0, 0, c.Controls.Flavor.Header, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignRight, false, 0, "")

		pdf.Ln(5)
	})

	pdf.SetFooterFunc(func() {
		pdf.SetFont(c.Controls.Pdf.Fonts.Footer, FontStyleItalic, 8)

		pdf.SetY(-15)

		footer := fmt.Sprintf("%s_%s_p%d",
			c.Controls.Flavor.Footer,
			time.Now().Format("2006-01-02-15-04-05-0700"),
			pdf.PageNo())

		hasher := sha1.New()
		io.WriteString(hasher, footer)
		hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

		hashWidth := pdf.GetStringWidth(hash)
		footerWidth := pdf.GetStringWidth(hash)

		pad := (WorkingPageWidth - hashWidth - footerWidth)

		pdf.Cell(hashWidth, 8, hash)
		pdf.CellFormat((footerWidth + pad), 8, footer, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignRight, false, 0, "")
	})

	width, _ := pdf.GetPageSize()
	leftMargin, _, rightMargin, _ := pdf.GetMargins()

	WorkingPageWidth = (width - leftMargin - rightMargin)
	DefaultFont = c.Controls.Pdf.Fonts.Default

	return pdf
}

func pdfContactLine(pdf *gofpdf.Fpdf, c *Configuration) {
	fontSize := float64(9)

	phoneNumberReplacer := strings.NewReplacer(
		"+", "",
		" ", "",
		"(", "",
		")", "",
		"-", "",
	)

	pdf.SetFont(DefaultFont, FontStyleNormal, fontSize)

	emailAddressLength := pdf.GetStringWidth(c.Contact.EmailAddress)
	phoneNumberLength := pdf.GetStringWidth(c.Contact.PhoneNumber)
	urlLength := pdf.GetStringWidth(c.Contact.Url)
	locationLength := pdf.GetStringWidth(c.Contact.Location)
	pad := ((WorkingPageWidth - emailAddressLength - phoneNumberLength - urlLength - locationLength) / 3)

	pdf.CellFormat(emailAddressLength, fontSize, c.Contact.EmailAddress, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignLeft, false, 0, "mailto:"+c.Contact.EmailAddress)
	pdf.CellFormat((phoneNumberLength + pad), fontSize, c.Contact.PhoneNumber, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignRight, false, 0, "tel:"+phoneNumberReplacer.Replace(c.Contact.PhoneNumber))
	pdf.CellFormat((urlLength + pad), fontSize, c.Contact.Url, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignRight, false, 0, c.Contact.Url)
	pdf.CellFormat((locationLength + pad), fontSize, c.Contact.Location, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignRight, false, 0, "")
}

func pdfSectionTitle(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFont(DefaultFont, FontStyleBold, 14)

	pdf.Ln(11)
	pdf.SetFillColor(200, 200, 200)
	pdf.SetCellMargin(1)
	pdf.Bookmark(title, 0, -1)
	pdf.CellFormat(0, 8.5, title, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignLeft, true, 0, "")
}

func pdfSkillsSection(pdf *gofpdf.Fpdf, cs *[]ConfigurationSkills, control *ConfigurationControlCountTagged) {
	if (control.Count == 0) || (len(*cs) == 0) {
		return
	}

	pdfSectionTitle(pdf, control.Title)

	fontSize := float64(11)
	lineBreak := (fontSize / 2)

	pdf.SetFont(DefaultFont, FontStyleNormal, fontSize)

	pdf.Ln(8)

	skills := make([]string, 0)
	skillsCount := uint(0)
	linesCount := uint(0)

skills_loop:
	// Iterate through all of the skills in the configuration
	for csi, css := range *cs {
		if css.Used {
			continue
		}

		var addSkill = func() bool {
			(*cs)[csi].Used = true

			skills = append(skills, css.Name)
			skillsCount++

			// Write a line if the *next* skill would be too wide
			if pdf.GetStringWidth(strings.Join(skills, " / ")) > WorkingPageWidth {
				if linesCount > 0 {
					pdf.Ln(lineBreak)
				}

				pdf.CellFormat(0, fontSize, strings.Join(skills[0:len(skills)-1], " / "), gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignCenter, false, 0, "")

				skills = []string{css.Name}

				linesCount++
			}

			return skillsCount == control.Count
		}

		if len(css.Tags) == 0 {
			if addSkill() {
				break skills_loop
			}
		} else {
		controls_tags_loop:
			// Iterate through this skill's tags
			for _, skillTag := range css.Tags {
				// Iterate through the tags we're targeting
				for _, controlTag := range control.Tags {
					if skillTag == controlTag {
						if addSkill() {
							break skills_loop
						}

						break controls_tags_loop
					}
				}
			}
		}
	}

	// Just in case we have leftovers
	if len(skills) > 0 {
		if linesCount > 0 {
			pdf.Ln(lineBreak)
		}

		pdf.CellFormat(0, fontSize, strings.Join(skills, " / "), gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignCenter, false, 0, "")
	}
}

func pdfOrganizationalExperience(pdf *gofpdf.Fpdf, co *[]ConfigurationOrganization, control *ConfigurationControlsOrganizations, condensed bool) {
	var sectionTitle string

	if condensed {
		if (control.Condensed.Count == 0) || (len(*co) == 0) {
			return
		}

		sectionTitle = control.Condensed.Title
	} else {
		if (control.Expanded.Count == 0) || (len(*co) == 0) {
			return
		}

		sectionTitle = control.Expanded.Title
	}

	pdfSectionTitle(pdf, sectionTitle)

	bulletCellWidth := float64(7)
	bulletPointWidth := (WorkingPageWidth - bulletCellWidth)

	organizationsCount := 0
	maxBulletPointsCount := control.Expanded.BulletPoints.Start

	var singleOrganizationHeightGuideline float64
	organizationNeedsNewline := true

organizations_loop:
	for coi, organization := range *co {
		if organization.Used {
			continue
		}

		var addOrganization = func() bool {
			(*co)[coi].Used = true

			if organizationsCount == 0 {
				singleOrganizationHeightGuideline = pdf.GetY()
			}

			// Line 1
			if organizationNeedsNewline {
				pdf.Ln(8)
			}

			fontSize := float64(11)
			lineBreak := fontSize

			// We need a single break when collapsing to the first position only
			if (len(organization.Positions) == 1) || (control.Expanded.CollapseMultiplePositions != CollapseMultiplePositionsFull) {
				lineBreak /= 2
			}

			pdf.SetFont(DefaultFont, FontStyleItalic, fontSize)

			var organizationExtra string

			if organization.OrganizationExtra != "" {
				organizationExtra = fmt.Sprintf(" (%s)", organization.OrganizationExtra)
			}

			organizationWidth := pdf.GetStringWidth(organization.Organization)
			organizationExtraWidth := pdf.GetStringWidth(organizationExtra)

			fontSize = float64(10)

			pdf.SetFontSize(fontSize)

			locationWidth := pdf.GetStringWidth(organization.Location)

			pad := (WorkingPageWidth - organizationWidth - organizationExtraWidth - locationWidth)

			fontSize = float64(11)

			pdf.SetFontSize(fontSize)

			pdf.Bookmark(organization.Organization, 1, -1)
			pdf.CellFormat(organizationWidth, fontSize, organization.Organization, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignLeft, false, 0, organization.Url)

			if organizationExtraWidth > 0 {
				pdf.Cell(organizationExtraWidth, fontSize, organizationExtra)
			}

			fontSize = float64(10)

			pdf.SetFontSize(fontSize)

			pdf.CellFormat((locationWidth + pad), fontSize, organization.Location, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignRight, false, 0, "")
			pdf.Ln(lineBreak)

			// Line 2: position start
			fontSize = float64(12)
			lineBreak = (fontSize / 2)

			pdf.SetFont(DefaultFont, FontStyleBold, fontSize)

			positionsListed := 0
			maxPositionIndex := (len(organization.Positions) - 1)

		positions_loop:
			for epi, position := range organization.Positions {
				if position.Used {
					continue
				}

				var addPosition = func() bool {
					var addPositionTitleLine = func(i int, p ConfigurationOrganizationPosition, renderLineBreak bool) {
						(*co)[coi].Positions[i].Used = true

						positionsListed++

						var title, flavor string

						if position.NormalizedTitle != "" {
							title = p.NormalizedTitle
						} else {
							title = p.Title
						}

						if p.Flavor != "" {
							flavor = fmt.Sprintf(" - %s", position.Flavor)
						}

						dates := fmt.Sprintf("%s to %s", p.Dates.Start, p.Dates.End)

						titleWidth := pdf.GetStringWidth(title)
						datesWidth := pdf.GetStringWidth(dates)

						pdf.SetFontStyle(FontStyleNormal)

						flavorWidth := pdf.GetStringWidth(flavor)

						pad := (WorkingPageWidth - titleWidth - flavorWidth - datesWidth)

						pdf.SetFontStyle(FontStyleBold)
						pdf.Cell(titleWidth, fontSize, title)

						if flavorWidth > 0 {
							pdf.SetFontStyle(FontStyleNormal)
							pdf.Cell(flavorWidth, fontSize, flavor)
						}

						pdf.SetFontStyle(FontStyleBold)
						pdf.CellFormat((datesWidth + pad), fontSize, dates, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignRight, false, 0, "")

						if !condensed && renderLineBreak {
							pdf.Ln(lineBreak)
						}
					}

					addPositionTitleLine(epi, position, ((epi < maxPositionIndex) || (len(position.Summary) > 0)))

					// If we're collapsing positions into titles only, then render them
					if control.Expanded.CollapseMultiplePositions == CollapseMultiplePositionsTitlesOnly {
						for epi2, position := range organization.Positions {
							if position.Used {
								continue
							}

							addPositionTitleLine(epi2, position, (epi2 < maxPositionIndex))
						}
					}

					if condensed {
						return true
					}

					// Possible line 3: position summary
					if position.Summary != "" {
						fontSize = float64(11)

						pdf.SetFont(DefaultFont, FontStyleNormal, fontSize)

						pdf.Cell(0, fontSize, position.Summary)

						pdf.Ln(fontSize)
					}

					// Lines 4+: bullet points
					fontSize = float64(11)
					lineBreak = (fontSize / 2)

					bulletPointsCount := uint(0)

					for bpi, bulletPoint := range position.BulletPoints {
						if bpi > 0 {
							pdf.Ln(lineBreak)
						}

						linesCount := 0
						bulletPointParts := strings.Split(bulletPoint, " ")

						for {
							var bullet string
							bulletPointsCollected := []string{}

							if linesCount == 0 {
								bullet = string(rune(117))
							} else {
								pdf.Ln(lineBreak)

								bullet = ""
							}

							pdf.SetFont("Symbol", FontStyleNormal, float64(6)) // Small bullets

							pdf.Cell(bulletCellWidth, fontSize, bullet)

							pdf.SetFont(DefaultFont, FontStyleNormal, fontSize)

							for _, bulletPointPart := range bulletPointParts {
								bulletPointsCollected = append(bulletPointsCollected, bulletPointPart)

								if pdf.GetStringWidth(strings.Join(bulletPointsCollected, " ")) > bulletPointWidth {
									sliceRight := len(bulletPointsCollected) - 1

									pdf.Cell(0, fontSize, strings.Join(bulletPointsCollected[0:sliceRight], " "))

									bulletPointParts = bulletPointParts[sliceRight:]
									bulletPointsCollected = []string{}
								}
							}

							bulletPointPartsCount := len(bulletPointParts)

							if (bulletPointPartsCount == 0) || (bulletPointPartsCount == len(bulletPointsCollected)) {
								pdf.Cell(0, fontSize, strings.Join(bulletPointsCollected, " "))

								break
							}

							linesCount++
						}

						// Check if we've hit the limit, and decrement for the next position
						bulletPointsCount++

						if bulletPointsCount == maxBulletPointsCount {
							maxBulletPointsCount -= control.Expanded.BulletPoints.Decrement

							break
						}
					}

					return control.Expanded.CollapseMultiplePositions == CollapseMultiplePositionsCollapse
				}

				if len(control.Expanded.PositionTags) == 0 {
					if addPosition() {
						break positions_loop
					}
				} else {
				controls_tags_loop:
					// Iterate through this position's tags
					for _, positionTag := range position.Tags {
						// Iterate through the tags we're targeting
						for _, controlTag := range control.Expanded.PositionTags {
							if positionTag == controlTag {
								if addPosition() {
									break positions_loop
								}

								break controls_tags_loop
							}
						}
					}
				}
			}

			pdf.Ln(3)

			organizationsCount++

			organizationNeedsNewline = true

			if organizationsCount == 1 {
				singleOrganizationHeightGuideline = (pdf.GetY() - singleOrganizationHeightGuideline)
			} else {
				y := pdf.GetY()
				_, _, _, bottom := pdf.GetMargins()
				_, height := pdf.GetPageSize()

				if (y + singleOrganizationHeightGuideline) > (height - bottom) {
					pdf.AddPage()

					organizationNeedsNewline = false
				}
			}

			if !condensed && (organizationsCount == int(control.Expanded.Count)) {
				return true
			} else if condensed && (organizationsCount == int(control.Condensed.Count)) {
				return true
			}

			return false
		}

		if len(control.Expanded.Tags) == 0 {
			if addOrganization() {
				break organizations_loop
			}
		} else {
		controls_tags_loop:
			// Iterate through this organization's tags
			for _, organizationTag := range organization.Tags {
				// Iterate through the tags we're targeting
				for _, controlTag := range control.Expanded.Tags {
					if organizationTag == controlTag {
						if addOrganization() {
							break organizations_loop
						}

						break controls_tags_loop
					}
				}
			}
		}
	}
}

func pdfEducation(pdf *gofpdf.Fpdf, c *Configuration) {
	if (c.Controls.Education.Count == 0) || (len(c.Education) == 0) {
		return
	}

	pdfSectionTitle(pdf, c.Controls.Education.Title)

	educationCount := 0

	pdf.Ln(8)

education_loop:
	for ei, education := range c.Education {
		if education.Used {
			continue
		}

		var addEducation = func() bool {
			c.Education[ei].Used = true

			fontSize := float64(11)
			lineBreak := (fontSize / 2)

			if ei > 0 {
				pdf.Ln(lineBreak)
			}

			pdf.SetFont(DefaultFont, FontStyleBold, fontSize)

			titleWidth := pdf.GetStringWidth(education.Title)

			fontSize = float64(10)

			pdf.SetFont(DefaultFont, FontStyleItalic, fontSize)

			institutionWidth := pdf.GetStringWidth(education.Institution)
			pad := (WorkingPageWidth - titleWidth - institutionWidth)

			pdf.Bookmark(education.Title, 1, -1)

			fontSize = float64(11)

			pdf.SetFont(DefaultFont, FontStyleBold, fontSize)

			pdf.CellFormat((titleWidth + pad), fontSize, education.Title, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignLeft, false, 0, education.Url)

			fontSize = float64(10)

			pdf.SetFont(DefaultFont, FontStyleItalic, fontSize)

			fontSize = float64(11) // need the cells to be the same height

			pdf.CellFormat(institutionWidth, fontSize, education.Institution, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignRight, false, 0, "")

			educationCount++

			return educationCount == int(c.Controls.Education.Count)
		}

		if len(c.Controls.Education.Tags) == 0 {
			if addEducation() {
				break education_loop
			}
		} else {
		controls_tags_loop:
			// Iterate through this education's tags
			for _, educationTag := range education.Tags {
				// Iterate through the tags we're targeting
				for _, controlTag := range c.Controls.Education.Tags {
					if educationTag == controlTag {
						if addEducation() {
							break education_loop
						}

						break controls_tags_loop
					}
				}
			}
		}
	}
}

func pdfProjects(pdf *gofpdf.Fpdf, c *Configuration) {
	if (c.Controls.Projects.Count == 0) || (len(c.Projects) == 0) {
		return
	}

	// todo find infinite loop / logic error in this function

	pdfSectionTitle(pdf, c.Controls.Projects.Title)

	bulletCellWidth := float64(7)
	bulletPointWidth := (WorkingPageWidth - bulletCellWidth)

	projectsCount := 0

	var singleProjectHeightGuideline float64
	projectNeedsNewline := true

projects_loop:
	for pi, project := range c.Projects {
		if project.Used {
			continue
		}

		var addProject = func() bool {
			c.Projects[pi].Used = true

			if projectsCount == 0 {
				singleProjectHeightGuideline = pdf.GetY()
			}

			// Line 1
			if projectNeedsNewline {
				pdf.Ln(8)
			}

			fontSize := float64(11)
			lineBreak := (fontSize / 2)

			pdf.SetFont(DefaultFont, FontStyleItalic, fontSize)

			titleWidth := pdf.GetStringWidth(project.Title)

			fontSize = float64(10)

			pdf.SetFontSize(fontSize)

			locationWidth := pdf.GetStringWidth(project.Location)

			pad := (WorkingPageWidth - titleWidth - locationWidth)

			fontSize = float64(11)

			pdf.SetFontSize(fontSize)

			pdf.Bookmark(project.Title, 1, -1)
			pdf.CellFormat(titleWidth, fontSize, project.Title, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignLeft, false, 0, project.Url)

			fontSize = float64(10)

			pdf.SetFontSize(fontSize)

			pdf.CellFormat((locationWidth + pad), fontSize, project.Location, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignRight, false, 0, "")
			pdf.Ln(lineBreak)

			// Line 2: role start
			fontSize = float64(12)
			lineBreak = (fontSize / 2)

			pdf.SetFont(DefaultFont, FontStyleBold, fontSize)

			dates := fmt.Sprintf("%s to %s", project.Dates.Start, project.Dates.End)

			roleWidth := pdf.GetStringWidth(project.Role)
			datesWidth := pdf.GetStringWidth(dates)

			pdf.SetFontStyle(FontStyleNormal)

			pad = (WorkingPageWidth - roleWidth - datesWidth)

			pdf.SetFontStyle(FontStyleBold)
			pdf.Cell(roleWidth, fontSize, project.Role)

			pdf.SetFontStyle(FontStyleBold)
			pdf.CellFormat((datesWidth + pad), fontSize, dates, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignRight, false, 0, "")

			pdf.Ln(lineBreak) // todo determine if necessary

			// Possible line 3: project summary
			if project.Summary != "" {
				fontSize = float64(11)

				pdf.SetFont(DefaultFont, FontStyleNormal, fontSize)

				pdf.Cell(0, fontSize, project.Summary)

				pdf.Ln(fontSize)
			}

			// Lines 4+: bullet points
			fontSize = float64(11)
			lineBreak = (fontSize / 2)

			for bpi, bulletPoint := range project.BulletPoints {
				if bpi > 0 {
					pdf.Ln(lineBreak)
				}

				linesCount := 0
				bulletPointParts := strings.Split(bulletPoint, " ")

				for {
					var bullet string
					bulletPointsCollected := []string{}

					if linesCount == 0 {
						bullet = string(rune(117))
					} else {
						pdf.Ln(lineBreak)

						bullet = ""
					}

					pdf.SetFont("Symbol", FontStyleNormal, float64(6)) // Small bullets

					pdf.Cell(bulletCellWidth, fontSize, bullet)

					pdf.SetFont(DefaultFont, FontStyleNormal, fontSize)

					for _, bulletPointPart := range bulletPointParts {
						bulletPointsCollected = append(bulletPointsCollected, bulletPointPart)

						if pdf.GetStringWidth(strings.Join(bulletPointsCollected, " ")) > bulletPointWidth {
							sliceRight := len(bulletPointsCollected) - 1

							pdf.Cell(0, fontSize, strings.Join(bulletPointsCollected[0:sliceRight], " "))

							bulletPointParts = bulletPointParts[sliceRight:]
							bulletPointsCollected = []string{}
						}
					}

					bulletPointPartsCount := len(bulletPointParts)

					if (bulletPointPartsCount == 0) || (bulletPointPartsCount == len(bulletPointsCollected)) {
						pdf.Cell(0, fontSize, strings.Join(bulletPointsCollected, " "))

						break
					}

					linesCount++
				}
			}

			pdf.Ln(3)

			projectsCount++

			projectNeedsNewline = true

			if projectsCount == 1 {
				singleProjectHeightGuideline = (pdf.GetY() - singleProjectHeightGuideline)
			} else {
				y := pdf.GetY()
				_, _, _, bottom := pdf.GetMargins()
				_, height := pdf.GetPageSize()

				if (y + singleProjectHeightGuideline) > (height - bottom) {
					pdf.AddPage()

					projectNeedsNewline = false
				}
			}

			return projectsCount == int(c.Controls.Projects.Count)
		}

		if len(c.Controls.Projects.Tags) == 0 {
			if addProject() {
				break projects_loop
			}
		} else {
		controls_tags_loop:
			// Iterate through this project's tags
			for _, projectTag := range project.Tags {
				// Iterate through the tags we're targeting
				for _, controlTag := range c.Controls.Projects.Tags {
					if projectTag == controlTag {
						if addProject() {
							break projects_loop
						}

						break controls_tags_loop
					}
				}
			}
		}
	}
}

func pdfCertifications(pdf *gofpdf.Fpdf, c *Configuration) {
	if (c.Controls.Certifications.Count == 0) || (len(c.Certifications) == 0) {
		return
	}

	pdfSectionTitle(pdf, c.Controls.Certifications.Title)

	certificationsCount := 0

	pdf.Ln(8)

certifications_loop:
	for ci, certification := range c.Certifications {
		if certification.Used {
			continue
		}

		var addCertification = func() bool {
			c.Certifications[ci].Used = true

			fontSize := float64(11)
			lineBreak := (fontSize / 2)

			if ci > 0 {
				pdf.Ln(lineBreak)
			}

			pdf.SetFont(DefaultFont, FontStyleBold, fontSize)

			titleWidth := pdf.GetStringWidth(certification.Certification)

			pdf.SetFontStyle(FontStyleNormal)

			dates := fmt.Sprintf(" (%s-%s)", certification.Dates.Start, certification.Dates.End)
			datesWidth := pdf.GetStringWidth(dates)

			fontSize = float64(10)

			pdf.SetFont(DefaultFont, FontStyleItalic, fontSize)

			institutionWidth := pdf.GetStringWidth(certification.Authority)
			pad := (WorkingPageWidth - titleWidth - datesWidth - institutionWidth)

			pdf.Bookmark(certification.Certification, 1, -1)

			fontSize = float64(11)

			pdf.SetFont(DefaultFont, FontStyleBold, fontSize)

			pdf.CellFormat(titleWidth, fontSize, certification.Certification, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignLeft, false, 0, certification.Url)

			pdf.SetFontStyle(FontStyleNormal)

			pdf.Cell(datesWidth, fontSize, dates)

			fontSize = float64(10)

			pdf.SetFont(DefaultFont, FontStyleItalic, fontSize)

			fontSize = float64(11) // need the cells to be the same height

			pdf.CellFormat((institutionWidth + pad), fontSize, certification.Authority, gofpdf.BorderNone, gofpdf.LineBreakNone, gofpdf.AlignRight, false, 0, "")

			certificationsCount++

			return certificationsCount == int(c.Controls.Certifications.Count)
		}

		if len(c.Controls.Certifications.Tags) == 0 {
			if addCertification() {
				break certifications_loop
			}
		} else {
		controls_tags_loop:
			// Iterate through this certification's tags
			for _, certificationTag := range certification.Tags {
				// Iterate through the tags we're targeting
				for _, controlTag := range c.Controls.Certifications.Tags {
					if certificationTag == controlTag {
						if addCertification() {
							break certifications_loop
						}

						break controls_tags_loop
					}
				}
			}
		}
	}
}

func writePdf(pdf *gofpdf.Fpdf, c *Configuration) {
	var outputFilename string

	if flagGeneratedPdf != "" {
		outputFilename = flagGeneratedPdf
	} else {
		outputFilename = c.Controls.Pdf.Filename
	}

	pdf.OutputFileAndClose(outputFilename)
}
