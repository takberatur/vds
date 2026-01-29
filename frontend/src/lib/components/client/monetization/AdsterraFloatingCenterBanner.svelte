<script lang="ts">
	import { modal } from '@/stores';
	import { AdsterraBanner } from '@/components/client/monetization/index';

	let {
		onclose,
		setting = $bindable(null)
	}: {
		open?: boolean;
		onclose?: () => void;
		setting?: SettingsValue | null;
	} = $props();

	let settingMonetization = $derived(setting?.MONETIZE);

	const hanldeClose = () => {
		onclose?.();
	};

	export function openModal() {
		modal.open({
			content: `
				<div class="relative">
					<div
							class="m-auto flex w-full max-w-md items-center justify-center p-4"
					>
							<AdsterraBanner
									ad={settingMonetization?.banner_rectangle_ad_code}
							/>
					</div>
				</div>
			`,
			size: 'lg',
			transparent: true,
			showCloseButton: true,
			closeButtonPosition: 'out-top-right',
			clickOutside: false
		});
	}
</script>
