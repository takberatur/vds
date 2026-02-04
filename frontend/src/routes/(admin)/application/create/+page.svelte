<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { MetaTags } from 'svelte-meta-tags';
	import { superForm } from 'sveltekit-superforms';
	import { handleSubmitLoading } from '@/stores';
	import { AdminSidebarLayout } from '@/components/admin';
	import { AppAlertDialog } from '@/components/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Spinner } from '@/components/ui/spinner/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Switch } from '@/components/ui/switch/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import Icon from '@iconify/svelte';
	import { localizeHref } from '@/paraglide/runtime';

	let { data } = $props();
	let metaTags = $derived(data.pageMetaTags);

	let errorMessage = $state<string | null>(null);

	// svelte-ignore state_referenced_locally
	const { form, errors, submitting, enhance } = superForm(data.form, {
		id: `create-application-form`,
		dataType: 'json',
		resetForm: false,
		onSubmit() {
			handleSubmitLoading(true);
			errorMessage = null;
		},
		async onUpdate(event) {
			if (event.result.type === 'failure') {
				errorMessage = event.result.data.message;
				return;
			}
			if (event.result.type === 'success') {
				handleSubmitLoading(false);
				await goto(localizeHref('/application'));
				await invalidateAll();
			}
		},
		onError(event) {
			errorMessage = event.result.error.message;
			handleSubmitLoading(false);
		}
	});

	function onPackageNameChange(val?: string) {
		if (!val) return;

		// validate with regex
		const packageNameRegex = /^[a-zA-Z0-9._-]+$/;
		if (!packageNameRegex.test(val)) {
			$errors.package_name = ['Invalid package name format'];
			return;
		}
	}
</script>

<MetaTags {...metaTags} />
<AdminSidebarLayout page="Create Application" user={data.user} setting={data.settings}>
	<div class="@container/main flex flex-col gap-4">
		<div class="flex-none px-4 py-4 sm:px-6">
			<div class="space-y-1">
				<h1 class="text-2xl font-bold tracking-tight sm:text-3xl">Create Application</h1>
				<p class="text-sm text-muted-foreground">Create a new application.</p>
			</div>
		</div>
		<div class="space-y-4 rounded-md border border-neutral-300 px-3 py-5 dark:border-neutral-700">
			<form method="POST" class="relative space-y-6" use:enhance>
				<Field.Group>
					<Field.Set>
						<Field.Field orientation="horizontal">
							<Field.Content>
								<Field.Label for="is_active">
									Enable Application : <Badge variant={$form.is_active ? 'default' : 'destructive'}>
										{$form.is_active ? 'Enabled' : 'Disabled'}
									</Badge>
								</Field.Label>
								<Field.Description>
									{$form.is_active
										? 'Application will be active and available for use.'
										: 'Application will be inactive and not available for use.'}
								</Field.Description>
							</Field.Content>
							<input type="hidden" name="is_active" value={$form.is_active} />
							<Switch
								id="is_active"
								bind:checked={$form.is_active}
								name="is_active"
								class="cursor-pointer"
								onCheckedChange={(val) => ($form.is_active = val)}
							/>
						</Field.Field>
						<Field.Field>
							<Field.Label for="name">Name</Field.Label>
							<div class="relative">
								<Icon
									icon="material-symbols:title"
									class="absolute top-1/2 left-3 -translate-y-1/2"
								/>
								<Input
									bind:value={$form.name}
									id="name"
									name="name"
									type="text"
									class="ps-10"
									placeholder="Enter application name"
									autocomplete="name"
									disabled={$submitting}
								/>
							</div>
							{#if $errors.name}
								<Field.Error>{$errors.name}</Field.Error>
							{/if}
						</Field.Field>
						<Field.Field>
							<Field.Label for="package_name">Package Name</Field.Label>
							<div class="relative">
								<Icon
									icon="material-symbols:package-2-outline-sharp"
									class="absolute top-1/2 left-3 -translate-y-1/2"
								/>
								<Input
									bind:value={$form.package_name}
									id="package_name"
									name="package_name"
									type="text"
									class="ps-10"
									placeholder="Enter application package name"
									autocomplete="on"
									disabled={$submitting}
									oninput={(e) => onPackageNameChange((e.target as HTMLInputElement).value)}
								/>
							</div>
							{#if $errors.package_name}
								<Field.Error>{$errors.package_name}</Field.Error>
							{/if}
						</Field.Field>
						<Field.Field>
							<Field.Label for="version">Version</Field.Label>
							<div class="relative">
								<Icon icon="ix:version-history" class="absolute top-1/2 left-3 -translate-y-1/2" />
								<Input
									bind:value={$form.version}
									id="version"
									name="version"
									type="text"
									class="ps-10"
									placeholder="Enter application version"
									autocomplete="on"
									disabled={$submitting}
								/>
							</div>
							{#if $errors.version}
								<Field.Error>{$errors.version}</Field.Error>
							{/if}
						</Field.Field>
						<Field.Field>
							<Field.Label for="platform">Platform</Field.Label>
							<div class="relative">
								<Icon
									icon="tabler:device-mobile-code"
									class="absolute top-1/2 left-3 -translate-y-1/2"
								/>
								<Input
									bind:value={$form.platform}
									id="platform"
									name="platform"
									type="text"
									class="ps-10"
									placeholder="Enter application platform"
									autocomplete="on"
									disabled={$submitting}
								/>
							</div>
							{#if $errors.platform}
								<Field.Error>{$errors.platform}</Field.Error>
							{/if}
						</Field.Field>
						<Field.Field orientation="horizontal">
							<Field.Content>
								<Field.Label for="enable_monetization">
									Enable Monetization : <Badge
										variant={$form.enable_monetization ? 'default' : 'destructive'}
									>
										{$form.enable_monetization ? 'Enabled' : 'Disabled'}
									</Badge>
								</Field.Label>
								<Field.Description>
									{$form.enable_monetization
										? 'Application will be monetized.'
										: 'Application will not be monetized.'}
								</Field.Description>
							</Field.Content>
							<input type="hidden" name="enable_monetization" value={$form.enable_monetization} />
							<Switch
								id="enable_monetization"
								bind:checked={$form.enable_monetization}
								name="enable_monetization"
								class="cursor-pointer"
								onCheckedChange={(val) => ($form.enable_monetization = val)}
							/>
						</Field.Field>

						<Card.Root class="w-full">
							<Card.Header>
								<Card.Title>AdMob Configuration</Card.Title>
								<Card.Description>Enter your AdMob configuration below.</Card.Description>
							</Card.Header>
							<Card.Content>
								<Field.Group>
									<Field.Set>
										<Field.Field orientation="horizontal">
											<Field.Content>
												<Field.Label for="enable_admob">
													Enable AdMob : <Badge
														variant={$form.enable_admob ? 'default' : 'destructive'}
													>
														{$form.enable_admob ? 'Enabled' : 'Disabled'}
													</Badge>
												</Field.Label>
												<Field.Description>
													{$form.enable_admob
														? 'Application will use AdMob for monetization.'
														: 'Application will not use AdMob for monetization.'}
												</Field.Description>
											</Field.Content>
											<input type="hidden" name="enable_admob" value={$form.enable_admob} />
											<Switch
												id="enable_admob"
												bind:checked={$form.enable_admob}
												name="enable_admob"
												class="cursor-pointer"
												onCheckedChange={(val) => ($form.enable_admob = val)}
											/>
										</Field.Field>
										{#if $form.enable_admob}
											<Field.Set>
												<Field.Field>
													<Field.Label for="admob_ad_unit_id">AdMob Ad Unit ID</Field.Label>
													<div class="relative">
														<Icon
															icon="mdi:code"
															class="absolute top-1/2 left-3 -translate-y-1/2"
														/>
														<Input
															bind:value={$form.admob_ad_unit_id}
															id="admob_ad_unit_id"
															name="admob_ad_unit_id"
															type="text"
															class="ps-10"
															placeholder="Enter AdMob Ad Unit ID"
															autocomplete="on"
															disabled={$submitting}
														/>
													</div>
													{#if $errors.admob_ad_unit_id}
														<Field.Error>{$errors.admob_ad_unit_id}</Field.Error>
													{/if}
												</Field.Field>
												<Field.Field>
													<Field.Label for="admob_banner_ad_unit_id"
														>AdMob Banner Ad Unit ID</Field.Label
													>
													<div class="relative">
														<Icon
															icon="mdi:code"
															class="absolute top-1/2 left-3 -translate-y-1/2"
														/>
														<Input
															bind:value={$form.admob_banner_ad_unit_id}
															id="admob_banner_ad_unit_id"
															name="admob_banner_ad_unit_id"
															type="text"
															class="ps-10"
															placeholder="Enter AdMob Banner Ad Unit ID"
															autocomplete="on"
															disabled={$submitting}
														/>
													</div>
													{#if $errors.admob_banner_ad_unit_id}
														<Field.Error>{$errors.admob_banner_ad_unit_id}</Field.Error>
													{/if}
												</Field.Field>
												<Field.Field>
													<Field.Label for="admob_interstitial_ad_unit_id"
														>AdMob Interstitial Ad Unit ID</Field.Label
													>
													<div class="relative">
														<Icon
															icon="mdi:code"
															class="absolute top-1/2 left-3 -translate-y-1/2"
														/>
														<Input
															bind:value={$form.admob_interstitial_ad_unit_id}
															id="admob_interstitial_ad_unit_id"
															name="admob_interstitial_ad_unit_id"
															type="text"
															class="ps-10"
															placeholder="Enter AdMob Interstitial Ad Unit ID"
															autocomplete="on"
															disabled={$submitting}
														/>
													</div>
													{#if $errors.admob_interstitial_ad_unit_id}
														<Field.Error>{$errors.admob_interstitial_ad_unit_id}</Field.Error>
													{/if}
												</Field.Field>
												<Field.Field>
													<Field.Label for="admob_native_ad_unit_id"
														>AdMob Native Ad Unit ID</Field.Label
													>
													<div class="relative">
														<Icon
															icon="mdi:code"
															class="absolute top-1/2 left-3 -translate-y-1/2"
														/>
														<Input
															bind:value={$form.admob_native_ad_unit_id}
															id="admob_native_ad_unit_id"
															name="admob_native_ad_unit_id"
															type="text"
															class="ps-10"
															placeholder="Enter AdMob Native Ad Unit ID"
															autocomplete="on"
															disabled={$submitting}
														/>
													</div>
													{#if $errors.admob_native_ad_unit_id}
														<Field.Error>{$errors.admob_native_ad_unit_id}</Field.Error>
													{/if}
												</Field.Field>
												<Field.Field>
													<Field.Label for="admob_rewarded_ad_unit_id"
														>AdMob Rewarded Ad Unit ID</Field.Label
													>
													<div class="relative">
														<Icon
															icon="mdi:code"
															class="absolute top-1/2 left-3 -translate-y-1/2"
														/>
														<Input
															bind:value={$form.admob_rewarded_ad_unit_id}
															id="admob_rewarded_ad_unit_id"
															name="admob_rewarded_ad_unit_id"
															type="text"
															class="ps-10"
															placeholder="Enter AdMob Rewarded Ad Unit ID"
															autocomplete="on"
															disabled={$submitting}
														/>
													</div>
													{#if $errors.admob_rewarded_ad_unit_id}
														<Field.Error>{$errors.admob_rewarded_ad_unit_id}</Field.Error>
													{/if}
												</Field.Field>
											</Field.Set>
										{/if}
									</Field.Set>
								</Field.Group>
							</Card.Content>
						</Card.Root>

						<Card.Root class="w-full">
							<Card.Header>
								<Card.Title>Unity Ad Configuration</Card.Title>
								<Card.Description>Enter your Unity Ad configuration below.</Card.Description>
							</Card.Header>
							<Card.Content>
								<Field.Group>
									<Field.Set>
										<Field.Field orientation="horizontal">
											<Field.Content>
												<Field.Label for="enable_unity_ad">
													Enable Unity Ad : <Badge
														variant={$form.enable_unity_ad ? 'default' : 'destructive'}
													>
														{$form.enable_unity_ad ? 'Enabled' : 'Disabled'}
													</Badge>
												</Field.Label>
												<Field.Description>
													{$form.enable_unity_ad
														? 'Application will use Unity Ad for monetization.'
														: 'Application will not use Unity Ad for monetization.'}
												</Field.Description>
											</Field.Content>
											<input type="hidden" name="enable_unity_ad" value={$form.enable_unity_ad} />
											<Switch
												id="enable_unity_ad"
												bind:checked={$form.enable_unity_ad}
												name="enable_unity_ad"
												class="cursor-pointer"
												onCheckedChange={(val) => ($form.enable_unity_ad = val)}
											/>
										</Field.Field>
										{#if $form.enable_unity_ad}
											<Field.Set>
												<Field.Field>
													<Field.Label for="unity_ad_unit_id">Unity Ad Unit ID</Field.Label>
													<div class="relative">
														<Icon
															icon="mdi:code"
															class="absolute top-1/2 left-3 -translate-y-1/2"
														/>
														<Input
															bind:value={$form.unity_ad_unit_id}
															id="unity_ad_unit_id"
															name="unity_ad_unit_id"
															type="text"
															class="ps-10"
															placeholder="Enter Unity Ad Unit ID"
															autocomplete="on"
															disabled={$submitting}
														/>
													</div>
													{#if $errors.unity_ad_unit_id}
														<Field.Error>{$errors.unity_ad_unit_id}</Field.Error>
													{/if}
												</Field.Field>
												<Field.Field>
													<Field.Label for="unity_banner_ad_unit_id"
														>Unity Banner Ad Unit ID</Field.Label
													>
													<div class="relative">
														<Icon
															icon="mdi:code"
															class="absolute top-1/2 left-3 -translate-y-1/2"
														/>
														<Input
															bind:value={$form.unity_banner_ad_unit_id}
															id="unity_banner_ad_unit_id"
															name="unity_banner_ad_unit_id"
															type="text"
															class="ps-10"
															placeholder="Enter Unity Banner Ad Unit ID"
															autocomplete="on"
															disabled={$submitting}
														/>
													</div>
													{#if $errors.unity_banner_ad_unit_id}
														<Field.Error>{$errors.unity_banner_ad_unit_id}</Field.Error>
													{/if}
												</Field.Field>
												<Field.Field>
													<Field.Label for="unity_interstitial_ad_unit_id"
														>Unity Interstitial Ad Unit ID</Field.Label
													>
													<div class="relative">
														<Icon
															icon="mdi:code"
															class="absolute top-1/2 left-3 -translate-y-1/2"
														/>
														<Input
															bind:value={$form.unity_interstitial_ad_unit_id}
															id="unity_interstitial_ad_unit_id"
															name="unity_interstitial_ad_unit_id"
															type="text"
															class="ps-10"
															placeholder="Enter Unity Interstitial Ad Unit ID"
															autocomplete="on"
															disabled={$submitting}
														/>
													</div>
													{#if $errors.unity_interstitial_ad_unit_id}
														<Field.Error>{$errors.unity_interstitial_ad_unit_id}</Field.Error>
													{/if}
												</Field.Field>
												<Field.Field>
													<Field.Label for="unity_rewarded_ad_unit_id"
														>Unity Rewarded Ad Unit ID</Field.Label
													>
													<div class="relative">
														<Icon
															icon="mdi:code"
															class="absolute top-1/2 left-3 -translate-y-1/2"
														/>
														<Input
															bind:value={$form.unity_rewarded_ad_unit_id}
															id="unity_rewarded_ad_unit_id"
															name="unity_rewarded_ad_unit_id"
															type="text"
															class="ps-10"
															placeholder="Enter Unity Rewarded Ad Unit ID"
															autocomplete="on"
															disabled={$submitting}
														/>
													</div>
													{#if $errors.unity_rewarded_ad_unit_id}
														<Field.Error>{$errors.unity_rewarded_ad_unit_id}</Field.Error>
													{/if}
												</Field.Field>
											</Field.Set>
										{/if}
									</Field.Set>
								</Field.Group>
							</Card.Content>
						</Card.Root>

						<Card.Root class="w-full">
							<Card.Header>
								<Card.Title>Start App Ad Configuration</Card.Title>
								<Card.Description>Enter your Start App Ad configuration below.</Card.Description>
							</Card.Header>
							<Card.Content>
								<Field.Group>
									<Field.Set>
										<Field.Field orientation="horizontal">
											<Field.Content>
												<Field.Label for="enable_start_app">
													Enable Start App : <Badge
														variant={$form.enable_start_app ? 'default' : 'destructive'}
													>
														{$form.enable_start_app ? 'Enabled' : 'Disabled'}
													</Badge>
												</Field.Label>
												<Field.Description>
													{$form.enable_start_app
														? 'Application will use Start App for monetization.'
														: 'Application will not use Start App for monetization.'}
												</Field.Description>
											</Field.Content>
											<input type="hidden" name="enable_start_app" value={$form.enable_start_app} />
											<Switch
												id="enable_start_app"
												bind:checked={$form.enable_start_app}
												name="enable_start_app"
												class="cursor-pointer"
												onCheckedChange={(val) => ($form.enable_start_app = val)}
											/>
										</Field.Field>
										{#if $form.enable_start_app}
											<Field.Set>
												<Field.Field>
													<Field.Label for="start_app_ad_unit_id">Start App Ad Unit ID</Field.Label>
													<div class="relative">
														<Icon
															icon="mdi:code"
															class="absolute top-1/2 left-3 -translate-y-1/2"
														/>
														<Input
															bind:value={$form.start_app_ad_unit_id}
															id="start_app_ad_unit_id"
															name="start_app_ad_unit_id"
															type="text"
															class="ps-10"
															placeholder="Enter Start App Ad Unit ID"
															autocomplete="on"
															disabled={$submitting}
														/>
													</div>
													{#if $errors.start_app_ad_unit_id}
														<Field.Error>{$errors.start_app_ad_unit_id}</Field.Error>
													{/if}
												</Field.Field>
											</Field.Set>
										{/if}
									</Field.Set>
								</Field.Group>
							</Card.Content>
						</Card.Root>

						<Card.Root class="w-full">
							<Card.Header>
								<Card.Title>One Signal Configuration</Card.Title>
								<Card.Description>Enter your One Signal configuration below.</Card.Description>
							</Card.Header>
							<Card.Content>
								<Field.Group>
									<Field.Set>
										<Field.Field>
											<Field.Label for="one_signal_id">One Signal ID</Field.Label>
											<div class="relative">
												<Icon
													icon="icon-park-outline:signal-one"
													class="absolute top-1/2 left-3 -translate-y-1/2"
												/>
												<Input
													bind:value={$form.one_signal_id}
													id="one_signal_id"
													name="one_signal_id"
													type="text"
													class="ps-10"
													placeholder="Enter One Signal ID"
													autocomplete="on"
													disabled={$submitting}
												/>
											</div>
											{#if $errors.one_signal_id}
												<Field.Error>{$errors.one_signal_id}</Field.Error>
											{/if}
										</Field.Field>
									</Field.Set>
								</Field.Group>
							</Card.Content>
						</Card.Root>

						<Field.Field orientation="horizontal">
							<Field.Content>
								<Field.Label for="enable_in_app_purchase">
									Enable In App Purchase : <Badge
										variant={$form.enable_in_app_purchase ? 'default' : 'destructive'}
									>
										{$form.enable_in_app_purchase ? 'Enabled' : 'Disabled'}
									</Badge>
								</Field.Label>
								<Field.Description>
									{$form.enable_in_app_purchase
										? 'Application will use In App Purchase for monetization.'
										: 'Application will not use In App Purchase for monetization.'}
								</Field.Description>
							</Field.Content>
							<input
								type="hidden"
								name="enable_in_app_purchase"
								value={$form.enable_in_app_purchase}
							/>
							<Switch
								id="enable_in_app_purchase"
								bind:checked={$form.enable_in_app_purchase}
								name="enable_in_app_purchase"
								class="cursor-pointer"
								onCheckedChange={(val) => ($form.enable_in_app_purchase = val)}
							/>
						</Field.Field>
					</Field.Set>
					<Field.Field>
						<Button type="submit" disabled={$submitting} class="w-full">
							{#if $submitting}
								<Spinner class="mr-2" />
								Please wait
							{:else}
								Create application
							{/if}
						</Button>
					</Field.Field>
				</Field.Group>
			</form>
		</div>
	</div>
	{#if errorMessage}
		<AppAlertDialog
			open={true}
			title="Error"
			message={errorMessage}
			type="error"
			onclose={() => (errorMessage = null)}
		/>
	{/if}
</AdminSidebarLayout>
