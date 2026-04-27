# `src` Developer Guide

## Muc tieu
- Giam thoi gian onboard cho dev moi.
- Giu cau truc de doc, de tim, de fix.
- Dung mot kieu router duy nhat: Next.js App Router.

## Cau truc thu muc
- `app/`: App Router + app-level providers/layout.
- `components/`: reusable UI (`ui`, `layout`, `feedback`, `scenic`).
- `api/`: HTTP client cho backend.
- `lib/`: utility va route constants.
- `i18n/`: message dictionary.
- `types/`: shared type.
- `generated/`: file sinh tu script.

## Quy uoc quan trong
1. Khong hardcode URL route trong screen/component.
2. Dung route constants tai `lib/routes.ts`.
3. Route file trong `app/**/page.tsx` la noi dat logic page theo segment.
4. Component co dung hook/browser API phai co `"use client"`.
5. Bien moi truong uu tien `NEXT_PUBLIC_*` (co the giu fallback cu khi can migrate).
6. Khong giu route group/file rong hoac component khong duoc su dung.

## Pattern tao route moi
1. Them route key trong `lib/routes.ts`.
2. Tao `app/<segment>/page.tsx` va dat page logic truc tiep.
3. Neu co UI duoc tai su dung, tach sang `components/`.
4. Neu can link/redirect, import `ROUTES` thay vi ghi chuoi.

## Checklist maintain
1. Moi route phai co owner ro rang trong `app/<segment>/page.tsx` hoac redirect gon.
2. Khong tao 2 bo component cung muc dich (tranh duplicate layer).
3. Context provider chi them khi it nhat 1 route dang dung.
4. Truoc merge, chay type-check: `pnpm exec tsc --noEmit`.
