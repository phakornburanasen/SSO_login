# User Rules

## Purpose

This document defines the user's general preferences for working with AI coding assistants
on the **SSO Login — Permission Controller** project.

These preferences should be applied when they do not conflict with project requirements, security requirements, or explicit instructions in the current task.

---

## 1. Communication

* **สื่อสารเป็นภาษาไทยเป็นหลัก** — คำอธิบาย สรุปงาน ข้อความ error ที่แสดงให้ผู้ใช้เห็น
* Technical terms เช่น `endpoint`, `commit`, `migration`, `IP`, `CIDR`, `baseUrl` ใช้ภาษาอังกฤษตามปกติ
* คำอธิบายให้กระชับ ตรงประเด็น ไม่ใช้สำนวนฟุ่มเฟือย
* ถ้ามีทางเลือกในการ implement ให้ชี้แจงทางเลือกสั้น ๆ ก่อนลงมือทำ

---

## 2. Understanding the Task

ก่อนแก้ไขโค้ดต้อง:

* **อ่านและทำความเข้าใจบริบทอย่างละเอียดก่อนดำเนินการ** (read-before-edit)
* อ่านไฟล์ที่เกี่ยวข้องทั้งหมด ไม่ใช่แค่ไฟล์เดียว
* ตรวจสอบว่ามีฟังก์ชันการทำงานเดิมอยู่แล้วหรือไม่ ก่อนสร้างใหม่
* ดูโครงสร้างโปรเจกต์ สไตล์การเขียน และ dependency ที่ใช้อยู่ก่อน
* หลีกเลี่ยงการตั้งสมมติฐานที่อาจทำให้กระทบไฟล์อื่น

สำหรับงานที่ชัดเจนและตรงไปตรงมา สามารถทำได้เลยโดยไม่ต้องถาม

---

## 3. Coding Style

* **เคารพและคงรักษาสไตล์การเขียนโค้ดเดิมของโปรเจกต์** (preserve existing code style)
* ใช้ Go standard style: `gofmt`, `go vet`
* ใช้ `net/http` ของ standard library ตามแบบที่โปรเจกต์ใช้อยู่
* แยก concerns ชัดเจน: `model → store → service → httpapi`
* หลีกเลี่ยง abstraction ที่ไม่จำเป็น
* ไม่เพิ่ม dependency ใหม่โดยไม่จำเป็น
* ไม่แก้ไขโค้ดที่ทำงานได้อยู่แล้วโดยไม่มีเหตุผล

---

## 4. Changes

* แก้ไขให้น้อยที่สุดที่แก้ปัญหาได้
* **ไม่แก้ไขไฟล์ที่ไม่เกี่ยวข้อง**
* ไม่เปลี่ยน architecture ถ้าไม่จำเป็น
* รักษาพฤติกรรมเดิมไว้ ถ้าไม่ได้รับคำสั่งให้เปลี่ยน

สำหรับการเปลี่ยนแปลงที่ส่งผลกระทบ ให้อธิบาย:

1. อะไรจะเปลี่ยน
2. ทำไมต้องเปลี่ยน
3. ส่วนไหนที่อาจได้รับผลกระทบ

---

## 5. Debugging

เมื่อแก้ปัญหา:

1. ระบุปัญหาให้ชัด
2. หา root cause
3. อธิบาย root cause
4. เสนอวิธีแก้
5. แก้ไข
6. ทดสอบผลลัพธ์

ไม่ใช้ workaround หรือซ่อน error โดยไม่เข้าใจสาเหตุ

---

## 6. Database Safety

สำหรับการทำงานกับ database:

* ตรวจสอบ schema เดิมก่อนแก้
* ก่อน `UPDATE` / `DELETE` ตรวจสอบจำนวน record ที่จะกระทบ
* ตรวจสอบ `WHERE` condition ให้ดี
* ไม่ทำ destructive operation (`DROP`, `TRUNCATE`, bulk `DELETE`) โดยไม่ได้รับการยืนยัน
* ใช้ migration file เสมอ — ไม่แก้ schema ผ่าน psql โดยตรง

---

## 7. Git Safety

ไม่ทำ Git operation ที่อาจทำลายข้อมูลโดยไม่ได้รับการยืนยัน เช่น:

* `git reset --hard`
* `git clean`
* force push
* ลบ branch
* discard uncommitted changes
* rewrite shared history

รักษา uncommitted changes ของผู้ใช้ไว้เสมอ

---

## 8. Dependencies

* ใช้ dependency เดิมเป็นหลัก — ตอนนี้มีแค่ `github.com/jackc/pgx/v5`
* ไม่เพิ่ม dependency ใหม่โดยไม่จำเป็น
* ถ้าจำเป็นต้องเพิ่ม ให้อธิบายเหตุผลก่อน

---

## 9. Security

ห้าม:

* ใส่ password ใน source code
* ใส่ API key ใน source code
* ใส่ access token ใน source code
* commit secrets ลง Git
* log secrets หรือ sensitive info

ใช้ environment variables ผ่าน `.env` หรือ Windows Environment Variables

---

## 10. Testing

หลังแก้ไข:

* รัน `go vet ./...` และ `go build ./...` เสมอ
* รัน `go test ./...` ถ้ามี test
* ตรวจสอบ syntax / compilation error
* รายงานผลการทดสอบให้ผู้ใช้ทราบ

ถ้าทดสอบไม่ได้ ให้อธิบายเหตุผล

---

## 11. Documentation

เมื่อมีการเปลี่ยนแปลงที่กระทบ:

* Architecture
* Installation
* Configuration
* API behavior
* Development workflow
* Deployment

ให้อัพเดท documentation ที่เกี่ยวข้อง (README.md, docs/, backend/API.md, backend/README.md)

ไม่สร้าง documentation สำหรับการเปลี่ยนแปลงเล็ก ๆ ที่ไม่จำเป็น

---

## 12. Reporting

เมื่อทำงานเสร็จ ให้สรุป:

1. **อะไรเปลี่ยน** — ไฟล์ที่ถูกแก้/สร้าง
2. **ทำไม** — เหตุผลของการเปลี่ยนแปลง
3. **ผลกระทบ** — ส่วนอื่นที่อาจได้รับผลกระทบ
4. **การทดสอบ** — ทดสอบอะไรไปบ้าง ผลเป็นอย่างไร
5. **ปัญหาที่เหลือ** — ถ้ามี

---

## 13. User Preference Overrides

คำสั่งเฉพาะของผู้ใช้ในงานปัจจุบันมีลำดับสูงกว่ากฎทั่วไปนี้ ยกเว้นเมื่อขัดกับ:

1. System / platform requirements
2. Security requirements
3. Mandatory project rules

ถ้ามีข้อขัดแย้ง ให้อธิบายสั้น ๆ และทำตามข้อที่สำคัญกว่า
